// vim:ts=4:sw=4:noexpandtab
package index

import (
	"github.com/Debian/dcs/cmd/dcs-web/common"
	"github.com/Debian/dcs/cmd/dcs-web/varz"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	varz.Increment("index-requests")
	err := common.Templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
