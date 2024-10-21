package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/auth"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/database"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/utilities"
	bolt "go.etcd.io/bbolt"
)

func getSetupForm(boltdb *bolt.DB) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		initialised, err := database.Initialised(boltdb)
		if err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to check if the application has been initialised: %w", err),
			)

			return
		}

		if initialised {
			sendClientError(
				writer,
				http.StatusForbidden,
				errors.New("the application is already initialised"),
			)

			return
		}

		form := formProfile{
			Website:  "",
			Password: "",
			Profile: formProfileInfo{
				DisplayName: "",
				URL:         "",
				Email:       "",
				PhotoURL:    "",
			},
			FieldErrors: make(map[string]string),
		}

		generateAndSendHTMLResponse(
			writer,
			"templates/html/setup.html.gotmpl",
			http.StatusOK,
			form,
		)
	})
}

func setupAccount(boltdb *bolt.DB) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		initialised, err := database.Initialised(boltdb)
		if err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to check if the application has been initialised: %w", err),
			)

			return
		}

		if initialised {
			sendClientError(
				writer,
				http.StatusForbidden,
				errors.New("the application is already initialised"),
			)

			return
		}

		if err := request.ParseForm(); err != nil {
			sendClientError(
				writer,
				http.StatusBadRequest,
				fmt.Errorf("error parsing the form: %w", err),
			)

			return
		}

		form := formProfile{
			Website:  request.PostFormValue("website"),
			Password: request.PostFormValue("password"),
			Profile: formProfileInfo{
				DisplayName: request.PostFormValue("profileDisplayName"),
				URL:         request.PostFormValue("profileURL"),
				PhotoURL:    request.PostFormValue("profilePhotoURL"),
				Email:       request.PostFormValue("profileEmail"),
			},
			FieldErrors: make(map[string]string),
		}

		if validForm := form.validate(); !validForm {
			generateAndSendHTMLResponse(
				writer,
				"templates/html/setup.html.gotmpl",
				http.StatusUnprocessableEntity,
				form,
			)

			return
		}

		hashedPassword, err := auth.HashPassword(form.Password)
		if err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to hash the password: %w", err),
			)
		}

		timestamp := time.Now()
		newProfile := database.Profile{
			CreatedAt:      timestamp,
			UpdatedAt:      timestamp,
			HashedPassword: hashedPassword,
			Information: database.ProfileInformation{
				Name:     form.Profile.DisplayName,
				URL:      form.Profile.URL,
				PhotoURL: form.Profile.PhotoURL,
				Email:    form.Profile.Email,
			},
		}

		if err := database.UpdateProfile(boltdb, form.Website, newProfile); err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to create the profile in the database: %w", err),
			)

			return
		}

		if err := database.UpdateInitialisedKey(boltdb); err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to update the app's initialised flag: %w", err),
			)
		}

		http.Redirect(writer, request, "/setup/confirmation", http.StatusSeeOther)
	})
}

func confirmation(writer http.ResponseWriter, _ *http.Request) {
	generateAndSendHTMLResponse(
		writer,
		"templates/html/confirmation.html.gotmpl",
		http.StatusOK,
		struct{}{},
	)
}

type formProfile struct {
	Website     string
	Password    string
	Profile     formProfileInfo
	FieldErrors map[string]string
}

type formProfileInfo struct {
	DisplayName string
	URL         string
	Email       string
	PhotoURL    string
}

func (n *formProfile) validate() bool {
	n.FieldErrors = make(map[string]string)

	canonicalisedWebsite, err := utilities.ValidateProfileURL(strings.TrimSpace(n.Website))
	if err != nil {
		slog.Error("profile website validation failed", "error", err.Error())
		n.FieldErrors["Website"] = "Please enter a valid website"
	} else {
		n.Website = canonicalisedWebsite
	}

	if utf8.RuneCountInString(n.Password) < 8 {
		n.FieldErrors["Password"] = "The password must be at least 8 characters long"
	}

	return len(n.FieldErrors) == 0
}
