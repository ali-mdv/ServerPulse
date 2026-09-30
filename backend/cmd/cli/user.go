package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var validate = validator.New()

type userView struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func toUserView(u models.User) userView {
	return userView{
		ID:        u.ID.Hex(),
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newUserCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}
	cmd.AddCommand(
		newUserListCmd(a),
		newUserGetCmd(a),
		newUserCreateCmd(a),
		newUserUpdateCmd(a),
		newUserDeleteCmd(a),
	)
	return cmd
}

func newUserListCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all users",
		RunE: func(cmd *cobra.Command, args []string) error {
			users, err := a.users.UsersList()
			if err != nil {
				return err
			}
			views := make([]userView, 0, len(*users))
			for _, u := range *users {
				views = append(views, toUserView(u))
			}
			return printJSON(cmd.OutOrStdout(), views)
		},
	}
}

func newUserGetCmd(a *app) *cobra.Command {
	var id, email string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve a user by ID or email",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (id == "") == (email == "") {
				return errors.New("provide exactly one of --id or --email")
			}

			var user *models.User
			var err error
			if id != "" {
				user, err = a.users.FindUserByID(id)
			} else {
				user, err = a.users.FindUserByEmail(email)
			}
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					return errors.New("user not found")
				}
				return err
			}
			return printJSON(cmd.OutOrStdout(), toUserView(*user))
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "user ID")
	cmd.Flags().StringVar(&email, "email", "", "user email")
	return cmd
}

func newUserCreateCmd(a *app) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			dto := dtos.CreateUserDTO{Email: email, Password: password}
			if err := validate.Struct(dto); err != nil {
				return fmt.Errorf("invalid input: %w", err)
			}

			user, err := a.users.CreateUser(dto)
			if err != nil {
				if errors.Is(err, apperrors.ErrConflict) {
					return fmt.Errorf("user %s already exists", email)
				}
				return err
			}
			return printJSON(cmd.OutOrStdout(), toUserView(*user))
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "user email (required)")
	cmd.Flags().StringVar(&password, "password", "", "user password (required)")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func newUserUpdateCmd(a *app) *cobra.Command {
	var id, email, password string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a user's email and/or password",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == "" {
				return errors.New("--id is required")
			}

			dto := dtos.UpdateUserDTO{}
			if cmd.Flags().Changed("email") {
				dto.Email = &email
			}
			if cmd.Flags().Changed("password") {
				dto.Password = &password
			}
			if dto.Email == nil && dto.Password == nil {
				return errors.New("provide at least one of --email or --password")
			}
			if err := validate.Struct(dto); err != nil {
				return fmt.Errorf("invalid input: %w", err)
			}

			user, err := a.users.UpdateUserByID(id, dto)
			if err != nil {
				switch {
				case errors.Is(err, mongo.ErrNoDocuments):
					return errors.New("user not found")
				case errors.Is(err, apperrors.ErrConflict):
					return fmt.Errorf("email %s is already in use", email)
				default:
					return err
				}
			}
			return printJSON(cmd.OutOrStdout(), toUserView(*user))
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "user ID (required)")
	cmd.Flags().StringVar(&email, "email", "", "new email")
	cmd.Flags().StringVar(&password, "password", "", "new password")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newUserDeleteCmd(a *app) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.users.DeleteUserByID(id); err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					return errors.New("user not found")
				}
				return err
			}
			return printJSON(cmd.OutOrStdout(), map[string]string{
				"id":      id,
				"message": "user deleted",
			})
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "user ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
