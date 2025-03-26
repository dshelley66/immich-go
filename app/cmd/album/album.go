package album

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/simulot/immich-go/app"
	"github.com/simulot/immich-go/internal/fileevent"
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
	cmd.AddCommand(NewAlbumShareCommand(ctx, cmd, a))
	cmd.AddCommand(NewAlbumUnshareCommand(ctx, cmd, a))
	cmd.AddCommand(NewAlbumAddUserAll(ctx, cmd, a))
	cmd.AddCommand(NewAlbumRemoveUserAll(ctx, cmd, a))

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

func NewAlbumShareCommand(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share <albumID | regex> <userID | regex>",
		Short: "Share album(s) to user(s) with role",
		Args:  cobra.ExactArgs(2),
	}
	var role string
	cmd.Flags().StringVarP(&role, "role", "r", "viewer", "Role to assign to user within the album(s)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		// logic:
		// determine if args[0] is UUID
		// determine if args[1] is UUID
		//
		// if albumID is UUID:
		//  - skip getting album list and put album ID in album array
		//  - get album info (also validate the album exists)
		// else
		// - get a list of albums, add to array any matching the regex
		// if userID is UUID:
		//  - skip getting user list and put userID in user array
		//  - get user info (also validate the user exists)
		// else
		// - get a list of users, add to array any matching the regex
		// for each album in album array
		//	for each user in user array
		//	- check album info to see if it is already shared with user
		//	- if not, add user to album
		albumID := args[0]
		userID := args[1]

		app.Log().Message("Adding user %s to album %s with role %s", albumID, userID, role)
		err := app.Client().Immich.AddUserToAlbum(ctx, albumID, userID, role)
		if err != nil {
			return fmt.Errorf("can't add the user to the album: %w", err)
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
		albumID := args[0]
		userID := args[1]

		app.Log().Message("Removing user %s from album %s", albumID, userID)
		err := app.Client().Immich.RemoveUserFromAlbum(ctx, albumID, userID)
		if err != nil {
			return fmt.Errorf("can't remove the user to the album: %w", err)
		}
		return nil
	}

	return cmd
}

// not needed anymore
func NewAlbumAddUserAll(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adduserall <userID>",
		Short: "Add user to all albums",
		Args:  cobra.ExactArgs(1),
	}

	var role string
	cmd.Flags().StringVarP(&role, "role", "r", "viewer", "Role to assign to user within the album")

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		log := app.Log()
		if app.Jnl() == nil {
			app.SetJnl(fileevent.NewRecorder(app.Log().Logger))
			app.Jnl().SetLogger(app.Log().SetLogWriter(os.Stdout))
		}
		userID := args[0]

		log.Message("Add user %s to all albums", userID)

		serverAlbums, err := app.Client().Immich.GetAllAlbums(ctx)
		if err != nil {
			return fmt.Errorf("can't get the album list from the server: %w", err)
		}
		for _, al := range serverAlbums {
			log.Message("Adding user %s to album %s with role %s", userID, al.ID, role)
			err := app.Client().Immich.AddUserToAlbum(ctx, al.ID, userID, role)
			if err != nil {
				return fmt.Errorf("can't add the user to album %s: %w", al.AlbumName, err)
			}
		}

		return nil
	}

	return cmd
}

// not needed anymore
func NewAlbumRemoveUserAll(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "removeuserall <userID>",
		Short: "Remove user's access to all albums",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		log := app.Log()
		if app.Jnl() == nil {
			app.SetJnl(fileevent.NewRecorder(app.Log().Logger))
			app.Jnl().SetLogger(app.Log().SetLogWriter(os.Stdout))
		}
		userID := args[0]

		log.Message("Remove user %s from all albums", userID)

		serverAlbums, err := app.Client().Immich.GetAllAlbums(ctx)
		if err != nil {
			return fmt.Errorf("can't get the album list from the server: %w", err)
		}
		for _, al := range serverAlbums {
			log.Message("Removing user %s from album %s", userID, al.ID)
			err := app.Client().Immich.RemoveUserFromAlbum(ctx, al.ID, userID)
			if err != nil {
				return fmt.Errorf("can't remove the user from album %s: %w", al.AlbumName, err)
			}
		}

		return nil
	}

	return cmd
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
