package gitlab

// import (
// 	"context"
// 	"fmt"
// 	"net/http"

// 	"github.com/tiagoposse/go-identity-sync/config"
// )

// type gitlabProvider struct {
// 	client *http.Client
// 	org    string
// }

// type myTransport struct {
// 	token string
// 	url   string
// }

// func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
// 	req.RequestURI = fmt.Sprintf("%s/%s", t.url, req.RequestURI)
// 	return http.DefaultTransport.RoundTrip(req)
// }

// func NewGitlabProvider(ctx context.Context, cfg *config.GitlabConfig) (*gitlabProvider, error) {
// 	client := &http.Client{
// 		Transport: &myTransport{
// 			token: *cfg.Token.Value,
// 			url:   fmt.Sprintf("%s/api/v4", cfg.Url),
// 		},
// 	}

// 	return &gitlabProvider{
// 		client: client,
// 		org:    cfg.Organisation,
// 	}, nil
// }

// func (gl *gitlabProvider) SearchUsers(ctx context.Context, filter string) {
// 	// Create a request to fetch user information by username
// 	// req, err := http.NewRequest("GET", fmt.Sprintf("/users?username=%s", username), nil)
// 	// if err != nil {
// 	// 	fmt.Println("Error creating request:", err)
// 	// 	return
// 	// }

// 	// // Set the GitLab group access token for authentication
// 	// req.Header.Set("Private-Token", groupToken)

// 	// // Send the request
// 	// resp, err := client.Do(req)
// 	// if err != nil {
// 	// 	fmt.Println("Error sending request:", err)
// 	// 	return
// 	// }
// 	// defer resp.Body.Close()

// 	// // Read the response body
// 	// body, err := ioutil.ReadAll(resp.Body)
// 	// if err != nil {
// 	// 	fmt.Println("Error reading response:", err)
// 	// 	return
// 	// }

// 	// // Unmarshal the JSON response to extract user details
// 	// var users []map[string]interface{}
// 	// err = json.Unmarshal(body, &users)
// 	// if err != nil {
// 	// 	fmt.Println("Error decoding JSON:", err)
// 	// 	return
// 	// }

// 	// // Display user details
// 	// if len(users) > 0 {
// 	// 	user := users[0] // Assuming the first user in the list is the matching user
// 	// 	fmt.Println("User details:")
// 	// 	fmt.Printf("Username: %v\n", user["username"])
// 	// 	fmt.Printf("Name: %v\n", user["name"])
// 	// 	fmt.Printf("Email: %v\n", user["email"])
// 	// 	// Add more fields as needed
// 	// } else {
// 	// 	fmt.Println("User not found")
// 	// }
// }
