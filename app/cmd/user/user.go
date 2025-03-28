package user

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/simulot/immich-go/app"
	"github.com/simulot/immich-go/app/cmd/util"
	"github.com/simulot/immich-go/internal/formats"
	"github.com/spf13/cobra"
)

func NewUserCommand(ctx context.Context, a *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User management commands",
	}
	app.AddClientFlags(ctx, cmd, a, false)

	cmd.AddCommand(NewUserListCommand(ctx, cmd, a))
	cmd.AddCommand(NewUserGetCommand(ctx, cmd, a))

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		return errors.New("you must specify a subcommand to the user command")
	}
	return cmd
}

func NewUserListCommand(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		users, err := app.Client().Immich.GetAllUsers(ctx)
		if err != nil {
			return fmt.Errorf("can't get the user list from the server: %w", err)
		}

		table := formats.OutFormatForList(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Email", "Created At", "Updated At"})
		for _, u := range users {
			table.Append([]string{u.ID, u.Name, u.Email, u.CreatedAt.Local().Format("2006-01-02 15:04:05 MST"), u.UpdatedAt.Local().Format("2006-01-02 15:04:05 MST")})
		}
		table.Render()
		return nil
	}
	return cmd
}
func NewUserGetCommand(ctx context.Context, parent *cobra.Command, app *app.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <userId>",
		Short: "Get user details",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error { //nolint:contextcheck
		user, err := app.Client().Immich.GetUserInfo(ctx, args[0])
		if err != nil {
			return fmt.Errorf("can't get the user %s: %w", args[0], err)
		}

		output, err := util.PrettyPrint(user)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "%s\n", output)

		return nil
	}

	return cmd
}
