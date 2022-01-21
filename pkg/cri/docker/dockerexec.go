/*
Package docker
@Time : 2021/12/28 17:01
@Author : sunxy
@File : dockerexec
@description:
*/
package docker

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/system"
	"github.com/pkg/errors"
	"github.com/sxy/lianfang/pkg/common"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//  Exec
//  @Description: 执行容器里面的命令 docker exec -it xxxx command
//  @param containerID
//  @param command
//  @return outPut
//  @return err
//
func Exec(containerID string, command []string) (outPut, errOutput string, err error) {
	outPut = ""
	ctx := context.Background()
	execOpts := types.ExecConfig{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	}
	execResp, err := common.DockerClient().ContainerExecCreate(ctx, containerID, execOpts)
	if err != nil {
		err = errors.Wrapf(err, "Create container [ %s ] exec failed", containerID)
		return
	}
	attach, err := common.DockerClient().ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		err = errors.Wrapf(err, "Attach container [ %s ] exec failed", containerID)
		return
	}
	defer attach.Close()
	outStream := new(bytes.Buffer)
	errStream := new(bytes.Buffer)
	_, err = stdcopy.StdCopy(outStream, errStream, attach.Reader)
	if err != nil {
		err = errors.Wrapf(err, "Read container [ %s ] exec result failed", containerID)
		return
	}
	outPut = strings.TrimSpace(outStream.String())
	errOutput = strings.TrimSpace(errStream.String())
	logger.Debug("container [ %s ], command %v, result: \n %s \n errout: \n %s \n", containerID, command, outPut, errOutput)
	return
}

//
//  CopyFrom
//  @Description:         copy file/dir from container to host path
//  @param containerId
//  @param srcFilePath    file/dir path in container
//  @param destFilePath   host path that file/dir will be copied to
//  @return output        copy command output
//  @return err           error
func CopyFrom(containerId, srcPath, dstPath string) (err error) {
	// cmd := util.BashCommand(fmt.Sprintf("docker cp %s:%s %s", containerId, srcFilePath, destFilePath))
	// var errOut string
	// output, errOut, err = util.ExecShell(cmd)
	// if err != nil || len(errOut) != 0 {
	// 	err = errors.Wrapf(err, fmt.Sprintf("[ %v ] failed, errOut is %s.", cmd, errOut))
	// 	return
	// }
	// return
	var rebaseName string
	srcStat, err := common.DockerClient().ContainerStatPath(context.Background(), containerId, srcPath)
	// If the destination is a symbolic link, we should follow it.
	if err == nil && srcStat.Mode&os.ModeSymlink != 0 {
		linkTarget := srcStat.LinkTarget
		if !system.IsAbs(linkTarget) {
			// Join with the parent directory.
			srcParent, _ := archive.SplitPathDirEntry(srcPath)
			linkTarget = filepath.Join(srcParent, linkTarget)
		}

		linkTarget, rebaseName = archive.GetRebaseName(srcPath, linkTarget)
		srcPath = linkTarget
	}
	content, stat, err := common.DockerClient().CopyFromContainer(context.Background(), containerId, srcPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer content.Close()

	srcInfo := archive.CopyInfo{
		Path:       srcPath,
		Exists:     true,
		IsDir:      stat.Mode.IsDir(),
		RebaseName: rebaseName,
	}

	preArchive := content
	if len(srcInfo.RebaseName) != 0 {
		_, srcBase := archive.SplitPathDirEntry(srcInfo.Path)
		preArchive = archive.RebaseArchiveEntries(content, srcBase, srcInfo.RebaseName)
	}
	err = archive.CopyTo(preArchive, srcInfo, dstPath)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

//
//  CopyFrom
//  @Description:         copy host file/dir to container
//  @param containerId
//  @param srcFilePath    file/dir path on host
//  @param destFilePath   container path that file/dir will be copied to
//  @return output        copy command output
//  @return err           error
func CopyTo(containerId, srcFilePath, destFilePath string) (err error) {
	return copyToContainer(containerId, srcFilePath, destFilePath, context.Background())
}

func copyToContainer(cid, srcPath, dstPath string, ctx context.Context) (err error) {
	if srcPath != "-" {
		// Get an absolute source path.
		srcPath, err = resolveLocalPath(srcPath)
		if err != nil {
			return err
		}
	}

	client := common.DockerClient()
	// Prepare destination copy info by stat-ing the container path.
	dstInfo := archive.CopyInfo{Path: dstPath}
	dstStat, err := client.ContainerStatPath(ctx, cid, dstPath)

	// If the destination is a symbolic link, we should evaluate it.
	if err == nil && dstStat.Mode&os.ModeSymlink != 0 {
		linkTarget := dstStat.LinkTarget
		if !system.IsAbs(linkTarget) {
			// Join with the parent directory.
			dstParent, _ := archive.SplitPathDirEntry(dstPath)
			linkTarget = filepath.Join(dstParent, linkTarget)
		}

		dstInfo.Path = linkTarget
		dstStat, err = client.ContainerStatPath(ctx, cid, linkTarget)
	}

	// Validate the destination path
	if err := ValidateOutputPathFileMode(dstStat.Mode); err != nil {
		return errors.Wrapf(err, `destination "%s:%s" must be a directory or a regular file`, cid, dstPath)
	}

	// Ignore any error and assume that the parent directory of the destination
	// path exists, in which case the copy may still succeed. If there is any
	// type of conflict (e.g., non-directory overwriting an existing directory
	// or vice versa) the extraction will fail. If the destination simply did
	// not exist, but the parent directory does, the extraction will still
	// succeed.
	if err == nil {
		dstInfo.Exists, dstInfo.IsDir = true, dstStat.Mode.IsDir()
	}

	var (
		content         io.Reader
		resolvedDstPath string
	)

	if srcPath == "-" {
		content = os.Stdin
		resolvedDstPath = dstInfo.Path
		if !dstInfo.IsDir {
			return errors.Errorf("destination \"%s:%s\" must be a directory", cid, dstPath)
		}
	} else {
		// Prepare source copy info.
		srcInfo, err := archive.CopyInfoSourcePath(srcPath, true)
		if err != nil {
			return err
		}

		srcArchive, err := archive.TarResource(srcInfo)
		if err != nil {
			return err
		}
		defer srcArchive.Close()

		// With the stat info about the local source as well as the
		// destination, we have enough information to know whether we need to
		// alter the archive that we upload so that when the server extracts
		// it to the specified directory in the container we get the desired
		// copy behavior.

		// See comments in the implementation of `archive.PrepareArchiveCopy`
		// for exactly what goes into deciding how and whether the source
		// archive needs to be altered for the correct copy behavior when it is
		// extracted. This function also infers from the source and destination
		// info which directory to extract to, which may be the parent of the
		// destination that the user specified.
		dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
		if err != nil {
			return err
		}
		defer preparedArchive.Close()

		resolvedDstPath = dstDir
		content = preparedArchive
	}

	options := types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
		CopyUIDGID:                false,
	}
	return client.CopyToContainer(ctx, cid, resolvedDstPath, content, options)
}

func resolveLocalPath(localPath string) (absPath string, err error) {
	if absPath, err = filepath.Abs(localPath); err != nil {
		return
	}
	return archive.PreserveTrailingDotOrSeparator(absPath, localPath, filepath.Separator), nil
}

// ValidateOutputPathFileMode validates the output paths of the `cp` command and serves as a
// helper to `ValidateOutputPath`
func ValidateOutputPathFileMode(fileMode os.FileMode) error {
	switch {
	case fileMode&os.ModeDevice != 0:
		return errors.New("got a device")
	case fileMode&os.ModeIrregular != 0:
		return errors.New("got an irregular file")
	}
	return nil
}
