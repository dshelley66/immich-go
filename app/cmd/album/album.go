package album

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/google/uuid"
	"github.com/simulot/immich-go/app"
	"github.com/simulot/immich-go/app/cmd/util"
	"github.com/simulot/immich-go/immich"
	"github.com/simulot/immich-go/internal/formats"
	"github.com/spf13/cobra"
)

func NewAlbumCommand(ctx context.Context, a *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "album",
		Short: "Album management commands",
	}
	app.AddClientFlags(ctx, cmd, a, false)

	cmd.AddCommand(NewAlbumListCommand(ctx, cmd, a))
	cmd.AddCommand(NewAlbumGetCommand(ctx, cmd, a))
	cmd.AddCommand(NewAlbumShareCommand(ctx, cmd, a))
	cmd.AddCommand(NewAlbumUnshareCommand(ctx, cmd, a))

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		return errors.New("you must specify a subcommand to the album command")
	}
	return cmd
}

func NewAlbumListCommand(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List albums",
	}
	var albumPattern string
	cmd.Flags().StringVarP(&albumPattern, "pattern", "p", ".*", "Pattern to filter the album list")

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		serverAlbums, err := app.Client().Immich.GetAllAlbums(ctx)
		if err != nil {
			return fmt.Errorf("can't get the album list from the server: %w", err)
		}

		table := formats.OutFormatForList(os.Stdout)
		table.SetHeader([]string{"ID", "Album Name", "Shared", "Shr Cnt", "Asset Count"})
		for _, al := range serverAlbums {
			if matched, _ := regexp.MatchString(albumPattern, al.AlbumName); matched {
				table.Append([]string{al.ID, al.AlbumName, fmt.Sprintf("%t", al.Shared), fmt.Sprintf("%d", len(al.AlbumUsers)),
					fmt.Sprintf("%d", al.AssetCount)})
			}
		}
		table.Render()
		return nil
	}
	return cmd
}

func NewAlbumGetCommand(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <albumID>",
		Short: "Get album details",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		album, err := app.Client().Immich.GetAlbumInfo(ctx, args[0], true)
		if err != nil {
			return fmt.Errorf("can't get the album %s: %w", args[0], err)
		}

		output, err := util.PrettyPrint(album)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "%s\n", output)

		return nil
	}
	return cmd
}

func NewAlbumShareCommand(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share <albumID | regex> <userID | regex>",
		Short: "Share album(s) to user(s) with role",
		Args:  cobra.ExactArgs(2),
	}
	var role string
	cmd.Flags().StringVarP(&role, "role", "r", "viewer", "Role to assign to user within the album(s)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		albumArray, err := getMatchingAlbums(ctx, app, args[0])
		if err != nil {
			return err
		}
		userArray, err := getMatchingUsers(ctx, app, args[1])
		if err != nil {
			return err
		}

		for _, album := range albumArray {
			for _, user := range userArray {
				if user.ID == album.Owner.ID {
					fmt.Printf("Album %s is owned by user %s...skipping\n", album.AlbumName, user.Name)
					continue
				}
				if !isAlbumSharedWithUser(&album, &user) {
					// Share the album with the user
					fmt.Printf("Adding user %s to album %s with role %s\n", user.Name, album.AlbumName, role)
					err := app.Client().Immich.AddUserToAlbum(ctx, album.ID, user.ID, role)
					if err != nil {
						return fmt.Errorf("can't add the user to the album: %w", err)
					}
				} else {
					fmt.Printf("Album %s is already shared with user %s...skipping\n", album.AlbumName, user.Name)
				}
			}
		}
		return nil
	}

	return cmd
}

func NewAlbumUnshareCommand(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unshare <albumID | regex> <userID | regex>",
		Short: "Remove user's access from album(s)",
		Args:  cobra.ExactArgs(2),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		albumArray, err := getMatchingAlbums(ctx, app, args[0])
		if err != nil {
			return err
		}
		userArray, err := getMatchingUsers(ctx, app, args[1])
		if err != nil {
			return err
		}

		for _, album := range albumArray {
			for _, user := range userArray {
				if user.ID == album.Owner.ID {
					fmt.Printf("Album %s is owned by user %s...skipping\n", album.AlbumName, user.Name)
					continue
				}
				if isAlbumSharedWithUser(&album, &user) {
					// Unshare the album from the user
					fmt.Printf("Unsharing user %s from album %s\n", user.Name, album.AlbumName)
					err := app.Client().Immich.RemoveUserFromAlbum(ctx, album.ID, user.ID)
					if err != nil {
						return fmt.Errorf("can't unshare the user from the album: %w", err)
					}
				} else {
					fmt.Printf("Album %s is not shared with user %s...skipping\n", album.AlbumName, user.Name)
				}
			}
		}
		return nil
	}

	return cmd
}

func isAlbumSharedWithUser(album *immich.AlbumSimplified, user *immich.User) bool {
	sharedWithUser := false
	if album.Shared {
		for _, sharedUser := range album.AlbumUsers {
			if sharedUser.User.ID == user.ID {
				sharedWithUser = true
				break
			}
		}
	}

	return sharedWithUser
}

func getMatchingAlbums(ctx context.Context, app *app.Application, albumIDorRegex string) ([]immich.AlbumSimplified, error) {
	var albumArray []immich.AlbumSimplified

	// Determine if albumID is UUID
	if _, err := uuid.Parse(albumIDorRegex); err == nil {
		// Skip getting album list and put album info in album array
		albumInfo, err := app.Client().Immich.GetAlbumInfo(ctx, albumIDorRegex, true)
		if err != nil {
			return albumArray, fmt.Errorf("failed to get album info for %s: %w", albumIDorRegex, err)
		}
		albumArray = append(albumArray, albumInfo)
	} else {
		// Get a list of albums and add to array any matching the regex
		albums, err := app.Client().Immich.GetAllAlbums(ctx)
		if err != nil {
			return albumArray, fmt.Errorf("failed to get all albums: %w", err)
		}
		for _, album := range albums {
			if match, _ := regexp.MatchString(albumIDorRegex, album.AlbumName); match {
				albumArray = append(albumArray, album)
			}
		}
	}
	if len(albumArray) == 0 {
		return albumArray, fmt.Errorf("no albums found matching %s", albumIDorRegex)
	}

	return albumArray, nil
}

func getMatchingUsers(ctx context.Context, app *app.Application, userIDorRegex string) ([]immich.User, error) {
	var userArray []immich.User

	// Determine if userID is UUID
	if _, err := uuid.Parse(userIDorRegex); err == nil {
		// Skip getting user list and put user info in user array
		userInfo, err := app.Client().Immich.GetUserInfo(ctx, userIDorRegex) // Assuming GetUserInfo exists
		if err != nil {
			return userArray, fmt.Errorf("failed to get user info for %s: %w", userIDorRegex, err)
		}
		userArray = append(userArray, userInfo)
	} else {
		// Get a list of users and add to array any matching the regex
		users, err := app.Client().Immich.GetAllUsers(ctx)
		if err != nil {
			return userArray, fmt.Errorf("failed to get users: %w", err)
		}
		for _, user := range users {
			if match, _ := regexp.MatchString(userIDorRegex, user.Name); match {
				userArray = append(userArray, user)
			}
		}
	}
	if len(userArray) == 0 {
		return userArray, fmt.Errorf("no users found matching %s", userIDorRegex)
	}

	return userArray, nil
}

// type DeleteAlbumCmd struct {
// 	*cmd.RootImmichFlags
// 	pattern   *regexp.Regexp // album pattern
// 	AssumeYes bool
// }

// func deleteAlbum(ctx context.Context, common *cmd.RootImmichFlags, args []string) error {
// 	app := &DeleteAlbumCmd{
// 		RootImmichFlags: common,
// 	}
// 	cmd := flag.NewFlagSet("album delete", flag.ExitOnError)
// 	app.RootImmichFlags.SetFlags(cmd)

// 	cmd.BoolFunc("yes", "When true, assume Yes to all actions", func(s string) error {
// 		var err error
// 		app.AssumeYes, err = strconv.ParseBool(s)
// 		return err
// 	})
// 	err := cmd.Parse(args)
// 	if err != nil {
// 		return err
// 	}
// 	err = app.RootImmichFlags.Start(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	args = cmd.Args()
// 	if len(args) > 0 {
// 		re, err := regexp.Compile(args[0])
// 		if err != nil {
// 			return fmt.Errorf("album pattern %q can't be parsed: %w", cmd.Arg(0), err)
// 		}
// 		app.pattern = re
// 	} else {
// 		app.pattern = regexp.MustCompile(`.*`)
// 	}

// 	albums, err := app.Immich.GetAllAlbums(ctx)
// 	if err != nil {
// 		return fmt.Errorf("can't get the albums list: %w", err)
// 	}
// 	sort.Slice(albums, func(i, j int) bool {
// 		return albums[i].AlbumName < albums[j].AlbumName
// 	})

// 	for _, al := range albums {
// 		if app.pattern.MatchString(al.AlbumName) {
// 			yes := app.AssumeYes
// 			if !yes {
// 				fmt.Printf("Delete album '%s'?\n", al.AlbumName)
// 				r, err := ui.ConfirmYesNo(ctx, "Proceed?", "n")
// 				if err != nil {
// 					return err
// 				}
// 				if r == "y" {
// 					yes = true
// 				}
// 			}
// 			if yes {
// 				fmt.Printf("Deleting album '%s'", al.AlbumName)
// 				err = app.Immich.DeleteAlbum(ctx, al.ID)
// 				if err != nil {
// 					return err
// 				} else {
// 					fmt.Println("done")
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }
