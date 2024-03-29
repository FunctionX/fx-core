package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/gorilla/mux"
)

const RestParamEvidenceHash = "evidence-hash"

// RegisterEvidenceRESTRoutes registers all Evidence submission handlers for the evidence module's
// REST service handler.
// Deprecated
func RegisterEvidenceRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	registerEvidenceQueryRoutes(clientCtx, WithHTTPDeprecationHeaders(rtr))
}

func registerEvidenceQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/evidence/{%s}", RestParamEvidenceHash), queryEvidenceHandler(clientCtx)).Methods(MethodGet)

	r.HandleFunc("/evidence", queryAllEvidenceHandler(clientCtx)).Methods(MethodGet)
}

func queryEvidenceHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		evidenceHash := vars[RestParamEvidenceHash]

		if strings.TrimSpace(evidenceHash) == "" {
			WriteErrorResponse(w, http.StatusBadRequest, "evidence hash required but not specified")
			return
		}

		clientCtx, ok := ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		decodedHash, err := hex.DecodeString(evidenceHash)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "invalid evidence hash")
			return
		}

		params := types.NewQueryEvidenceRequest(decodedHash)
		bz, err := clientCtx.Codec.MarshalJSON(params)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryEvidence)
		res, height, err := clientCtx.QueryWithData(route, bz)
		if CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		PostProcessResponse(w, clientCtx, res)
	}
}

func queryAllEvidenceHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := ParseHTTPArgsWithLimit(r, 0)
		if CheckBadRequestError(w, err) {
			return
		}

		clientCtx, ok := ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params := types.NewQueryAllEvidenceParams(page, limit)
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal query params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllEvidence)
		res, height, err := clientCtx.QueryWithData(route, bz)
		if CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		PostProcessResponse(w, clientCtx, res)
	}
}
