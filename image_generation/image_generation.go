package image_generation

import (
	"io"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/utils"
)

func ImageGeneration(w http.ResponseWriter, r *http.Request, _ router.Params) {
	prompt := r.URL.Query().Get("p")

	gen, err := Generate(prompt)
	if err != nil {
		utils.RespondError(w, "image_generation_error")
		return
	}

	reader, err := (gen.Data[0]).Reader()
	if err != nil {
		utils.RespondError(w, "read_error")
		return
	}
	io.Copy(w, reader)
}
