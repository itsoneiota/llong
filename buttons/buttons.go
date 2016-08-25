package buttons

import (
	"fmt"
	"net/http"

	"github.com/abourget/slick"
	"github.com/gorilla/mux"
)

type Buttons struct {
	bot *slick.Bot
}

func init() {
	slick.RegisterPlugin(&Buttons{})
}

func (buttons *Buttons) InitWebPlugin(bot *slick.Bot, privRouter *mux.Router, pubRouter *mux.Router) {
	buttons.bot = bot
	privRouter.HandleFunc("/slack/response", buttons.handleResponse)
	fmt.Println("Registered /slack/response")
}

func (buttons *Buttons) handleResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GOT A BUTTON RESPONSE")
	// w.Header().Set("Content-Type", "application/json")
	// enc := json.NewEncoder(w)
	// out := struct {
	// 	Users map[string]slack.User `json:"users"`
	// }{
	// 	Users: buttons.bot.Users,
	// }
	//
	// err := enc.Encode(out)
	// if err != nil {
	// 	webReportError(w, "Error encoding JSON", err)
	// 	return
	// }
	// return
}

func webReportError(w http.ResponseWriter, msg string, err error) {
	fmt.Println("BUTTON RESPONSE ERROR")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%s\n\n%s\n", msg, err)))
}
