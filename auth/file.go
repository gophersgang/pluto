package auth

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/numbleroot/pluto/utils"
)

// Structs

// FileAuthenticator contains file based authentication
// information including the in-memory map of username to
// password mapping.
type FileAuthenticator struct {
	lock      sync.Mutex
	File      string
	Separator string
	Users     []User
}

// User holds name and password from one line from users file.
type User struct {
	ID       int
	Name     string
	Password string
	Token    string
}

// Define list type of users to search efficiently.
type UsersByName []User

// Functions

// Make list of users searchable efficiently.
func (u UsersByName) Len() int           { return len(u) }
func (u UsersByName) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u UsersByName) Less(i, j int) bool { return u[i].Name < u[j].Name }

// NewFileAuthenticator takes in a file name and a separator,
// reads in specified file and parses it line by line as
// username - password elements separated by the separator.
// At the end, the returned struct contains the information
// and an in-memory map of username mapped to password.
func NewFileAuthenticator(file string, sep string) (*FileAuthenticator, error) {

	i := 1
	var err error
	var handle *os.File
	var nextUser User

	// Reserve space for the ordered users list in memory.
	Users := make([]User, 0, 50)

	// Open file with authentication information.
	handle, err = os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("[auth.NewFileAuthenticator] Could not open supplied authentication file: %s\n", err.Error())
	}
	defer handle.Close()

	// Create a new scanner on top of file handle.
	scanner := bufio.NewScanner(handle)

	// As long as there are lines left, scan them into memory.
	for scanner.Scan() {

		// Split read line based on separator defined in config file.
		userData := strings.Split(scanner.Text(), sep)

		// Create new user struct.
		nextUser = User{
			ID:       i,
			Name:     userData[0],
			Password: userData[1],
			Token:    "",
		}

		// Append new user element to slice.
		Users = append(Users, nextUser)

		// Increment original ID counter.
		i++
	}

	// If the scanner ended with an error, report it.
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("[auth.NewFileAuthenticator] Experienced error while scanning authentication file: %s\n", err.Error())
	}

	// Sort users list to search it efficiently later on.
	sort.Sort(UsersByName(Users))

	return &FileAuthenticator{
		lock:      sync.Mutex{},
		File:      file,
		Separator: sep,
		Users:     Users,
	}, nil
}

// GetIDOfUser finds position of supplied user in users
// list. It is assumed that existence check was already
// performed, for example via AuthenticatePlain.
func (f *FileAuthenticator) GetOriginalIDOfUser(username string) int {

	// This routine has to be safe for concurrent usage,
	// therefore lock the struct on entry.
	f.lock.Lock()
	defer f.lock.Unlock()

	// Search in user list for user matching supplied name.
	i := sort.Search(len(f.Users), func(i int) bool {
		return f.Users[i].Name >= username
	})

	return f.Users[i].ID
}

// GetTokenOfUser returns the currently assigned token as
// a sign of a valid authentication for a supplied name.
func (f *FileAuthenticator) GetTokenOfUser(username string) string {

	// This routine has to be safe for concurrent usage,
	// therefore lock the struct on entry.
	f.lock.Lock()
	defer f.lock.Unlock()

	// Search in user list for user matching supplied name.
	i := sort.Search(len(f.Users), func(i int) bool {
		return f.Users[i].Name >= username
	})

	// If that user does not exist, return the empty string.
	if !((i < len(f.Users)) && (f.Users[i].Name == username)) {
		return ""
	}

	// Take the token from user element and return it.
	return f.Users[i].Token
}

// DeleteTokenOfUser deletes the currently assigned token,
// logging the user out of the system.
func (f *FileAuthenticator) DeleteTokenOfUser(id int) {

	// This routine has to be safe for concurrent usage,
	// therefore lock the struct on entry.
	f.lock.Lock()
	defer f.lock.Unlock()

	// Set token to empty string.
	f.Users[id].Token = ""
}

// AuthenticatePlain performs the actual authentication
// process by taking supplied credentials and attempting
// to find a matching entry the in-memory list taken from
// the authentication file.
func (f *FileAuthenticator) AuthenticatePlain(username string, password string) (*int, *string, error) {

	// This routine has to be safe for concurrent usage,
	// therefore lock the struct on entry.
	f.lock.Lock()
	defer f.lock.Unlock()

	// Search in user list for user matching supplied name.
	i := sort.Search(len(f.Users), func(i int) bool {
		return f.Users[i].Name >= username
	})

	// If that user does not exist, throw an error.
	if !((i < len(f.Users)) && (f.Users[i].Name == username)) {
		return nil, nil, fmt.Errorf("Username not found in list of users")
	}

	// Check if a token has already present as a
	// sign of an earlier successful authentication.
	if token := f.Users[i].Token; token != "" {
		return &i, &token, nil
	}

	// Otherwise, check if passwords match.
	if f.Users[i].Password != password {
		return nil, nil, fmt.Errorf("Passwords did not match")
	}

	// Create new token and store it in element.
	newToken := utils.GenerateRandomString(32)
	f.Users[i].Token = newToken

	return &i, &newToken, nil
}