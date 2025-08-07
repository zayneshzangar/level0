package v1

import (
	"net/http"
	"order/config"

	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler, cfg *config.Config) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/order/{order_uid}", handler.GetOrder).Methods("GET")
	return r
}

func Cors(next http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Формируем правильный origin с http://
		origin := "http://" + cfg.Front.Host + ":" + cfg.Front.Port
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
