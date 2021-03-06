package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/RadhiFadlillah/boom/internal/fileutils"
	"github.com/RadhiFadlillah/boom/internal/model"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

func newSiteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "site [path]",
		Short: "Create a new site at specified path",
		Args:  cobra.MaximumNArgs(1),
		Run:   newSiteHandler,
	}

	cmd.Flags().StringP("title", "t", "", "title of the website")
	cmd.Flags().StringP("owner", "o", "", "owner of the website")
	cmd.Flags().BoolP("force", "f", false, "force init inside non-empty directory")
	return cmd
}

func newSiteHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	rootDir := "."
	if len(args) > 0 {
		rootDir = args[0]
	}

	title, _ := cmd.Flags().GetString("title")
	owner, _ := cmd.Flags().GetString("owner")
	isForced, _ := cmd.Flags().GetBool("force")

	title = strings.TrimSpace(title)
	owner = strings.TrimSpace(owner)

	// Make sure target directory exists
	os.MkdirAll(rootDir, os.ModePerm)

	// Make sure target dir is empty
	dirEmpty, err := fileutils.DirIsEmpty(rootDir)
	panicError(err)

	if !dirEmpty && !isForced {
		cError.Printf("Directory %s already exists and not empty\n", rootDir)
		return
	}

	// Ask metadata from user
	scanner := bufio.NewScanner(os.Stdin)
	if title == "" {
		cBold.Print("Website title : ")
		scanner.Scan()
		title = strings.TrimSpace(scanner.Text())

		if title == "" {
			cError.Println("Website title must not empty")
			return
		}
	}

	if owner == "" {
		cBold.Print("Website owner : ")
		scanner.Scan()
		owner = strings.TrimSpace(scanner.Text())
	}

	// Generate metadata
	bt, err := toml.Marshal(model.Metadata{
		Title:      title,
		Author:     owner,
		CreateTime: time.Now(),
		Pagination: 10})
	panicError(err, "Failed to create metadata:")

	// Create directories
	os.MkdirAll(fp.Join(rootDir, "themes"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "assets"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "content"), os.ModePerm)

	// Write index page
	indexBuffer := bytes.NewBuffer(nil)
	indexBuffer.WriteString("+++\n")
	indexBuffer.Write(bt)
	indexBuffer.WriteString("+++\n\n")
	indexBuffer.WriteString("Welcome to your site")

	prefixErrIndex := "Failed to create index page:"
	indexPath := fp.Join(rootDir, "content", "_index.md")
	indexFile, err := os.Create(indexPath)
	panicError(err, prefixErrIndex)
	defer indexFile.Close()

	_, err = io.Copy(indexFile, indexBuffer)
	panicError(err, prefixErrIndex)

	// Finish
	fmt.Print("Your new site is created in ")
	cBold.Println(rootDir)
}
