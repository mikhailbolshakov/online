package server

import (
	"chats/sdk"
	"encoding/json"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RoomHttpService struct {
	ws *WsServer
}

func (s *RoomHttpService) setRouting(router *mux.Router) {

	router.HandleFunc("/api/v1/rooms", func(writer http.ResponseWriter, request *http.Request) {
		s.Create(writer, request)
	}).Methods("POST")

	router.HandleFunc("/api/v1/rooms/subscribe", func(writer http.ResponseWriter, request *http.Request) {
		s.Subscribe(writer, request)
	}).Methods("POST")

	router.HandleFunc("/api/v1/rooms", func(writer http.ResponseWriter, request *http.Request) {
		s.GetRooms(writer, request)
	}).Methods("GET")

	router.HandleFunc("/api/v1/rooms/close", func(writer http.ResponseWriter, request *http.Request) {
		s.Close(writer, request)
	}).Methods("POST")

	router.HandleFunc("/api/v1/rooms/messages/history", func(writer http.ResponseWriter, request *http.Request) {
		s.GetMessageHistory(writer, request)
	}).Methods("GET")

}

func (s *RoomHttpService) Create(writer http.ResponseWriter, request *http.Request) {

	rq := &sdk.CreateRoomRequest{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(rq); err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "Invalid request payload")
		return
	}

	rs, err := s.ws.CreateRoom(rq)
	if err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, err.Message)
		return
	}

	s.ws.httpServer.respondWithJSON(writer, http.StatusCreated, rs)

}

func (s *RoomHttpService) Subscribe(writer http.ResponseWriter, request *http.Request) {

	rq := &sdk.RoomSubscribeRequest{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(rq); err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "Invalid request payload")
		return
	}

	rs, err := s.ws.RoomSubscribe(rq)
	if err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusInternalServerError, err.Message)
		return
	}

	s.ws.httpServer.respondWithJSON(writer, http.StatusCreated, rs)

}

func (s *RoomHttpService) GetRooms(writer http.ResponseWriter, request *http.Request) {

	rq := &sdk.GetRoomsByCriteriaRequest{
		AccountId:       &sdk.AccountIdRequest{
			ExternalId: request.FormValue("externalId"),
		},
		ReferenceId:     request.FormValue("referenceId"),
		RoomId:          uuid.UUID{},
		WithClosed:      false,
		WithSubscribers: true,
	}

	closedText := request.FormValue("closed")
	if closedText != "" {
		closed, e := strconv.ParseBool(closedText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "closed error: " + e.Error())
			return
		}
		rq.WithClosed = closed
	}


	if roomIdtext := request.FormValue("roomId"); roomIdtext != "" {
		roomId, e := uuid.FromString(roomIdtext)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "roomId error: "+e.Error())
			return
		}
		rq.RoomId = roomId
	}

	if accountIdtext := request.FormValue("accountId"); accountIdtext != "" {
		accountId, e := uuid.FromString(accountIdtext)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "accountId error: "+e.Error())
			return
		}
		rq.AccountId.AccountId = accountId
	}

	rs, err := s.ws.GetRoomsByCriteria(rq)
	if err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusInternalServerError, err.Message)
		return
	}

	s.ws.httpServer.respondWithJSON(writer, http.StatusOK, rs)

}

func (s *RoomHttpService) Close(writer http.ResponseWriter, request *http.Request) {

	rq := &sdk.CloseRoomRequest{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(rq); err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "Invalid request payload")
		return
	}

	rs, err := s.ws.CloseRoom(rq)
	if err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, err.Message)
		return
	}

	s.ws.httpServer.respondWithJSON(writer, http.StatusCreated, rs)

}

func (s *RoomHttpService) GetMessageHistory(writer http.ResponseWriter, request *http.Request) {

	rq := &sdk.GetMessageHistoryRequest{
		PagingRequest: &sdk.PagingRequest{
			SortBy: []sdk.SortRequest{},
		},
		Criteria:      &sdk.GetMessageHistoryCriteria{
			AccountId:     sdk.AccountIdRequest{},
		},
	}

	if roomIdtext := request.FormValue("roomId"); roomIdtext != "" {
		roomId, e := uuid.FromString(roomIdtext)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "roomId: "+e.Error())
			return
		}
		rq.Criteria.RoomId = roomId
	}

	if accountIdtext := request.FormValue("accountId"); accountIdtext != "" {
		accountId, e := uuid.FromString(accountIdtext)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "accountId: "+e.Error())
			return
		}
		rq.Criteria.AccountId.AccountId = accountId
	}

	rq.Criteria.AccountId.ExternalId = request.FormValue("externalId")
	rq.Criteria.ReferenceId = request.FormValue("referenceId")

	withStatusesText := request.FormValue("withStatuses")
	if withStatusesText != "" {
		withStatuses, e := strconv.ParseBool(withStatusesText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "withStatusesText: " + e.Error())
			return
		}
		rq.Criteria.WithStatuses = withStatuses
	}

	withAccountsText := request.FormValue("withAccounts")
	if withAccountsText != "" {
		withAccounts, e := strconv.ParseBool(withAccountsText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "withAccounts: " + e.Error())
			return
		}
		rq.Criteria.WithAccounts = withAccounts
	}

	sentOnlyText := request.FormValue("sentOnly")
	if sentOnlyText != "" {
		sentOnly, e := strconv.ParseBool(sentOnlyText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "sentOnly: " + e.Error())
			return
		}
		rq.Criteria.SentOnly = sentOnly
	}

	receivedOnlyText := request.FormValue("receivedOnly")
	if receivedOnlyText != "" {
		receivedOnly, e := strconv.ParseBool(receivedOnlyText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "receivedOnly: " + e.Error())
			return
		}
		rq.Criteria.ReceivedOnly = receivedOnly
	}

	createdBeforeText := request.FormValue("createdBefore")
	if createdBeforeText != "" {
		createdBefore, e := time.Parse(time.RFC3339, createdBeforeText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "createdBefore: " + e.Error())
			return
		}
		rq.Criteria.CreatedBefore = &createdBefore
	}

	createdAfterText := request.FormValue("createdAfter")
	if createdAfterText != "" {
		createdAfter, e := time.Parse(time.RFC3339, createdAfterText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "createdAfter: " + e.Error())
			return
		}
		rq.Criteria.CreatedAfter = &createdAfter
	}

	pageSizeText := request.FormValue("pageSize")
	if pageSizeText != "" {
		pageSize, e := strconv.Atoi(pageSizeText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "pageSize: " + e.Error())
			return
		}
		rq.PagingRequest.Size = pageSize
	}

	pageIndexText := request.FormValue("pageIndex")
	if pageIndexText != "" {
		pageIndex, e := strconv.Atoi(pageIndexText)
		if e != nil {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "pageIndex: " + e.Error())
			return
		}
		rq.PagingRequest.Index = pageIndex
	}

	sortBy := request.FormValue("sort")
	if sortBy != "" {
		sortExprSlice := strings.Split(sortBy, ",")
		if len(sortExprSlice) == 0 {
			s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "sortBy: incorrect format")
			return
		}
		for _, expr := range sortExprSlice {
			ops := strings.Split(expr, ":")
			if len(ops) != 2 {
				s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "sort: incorrect format")
				return
			}
			// TODO: currently only supported sort field
			// we imply sort parameters passed like sortBy=createdAt:asc,otherField:desc etc.
			if ops[0] != "createdAt" || (ops[1] != "asc" && ops[1] != "desc") {
				s.ws.httpServer.respondWithError(writer, http.StatusBadRequest, "sort: incorrect format")
				return
			}
			rq.PagingRequest.SortBy = append(rq.PagingRequest.SortBy, sdk.SortRequest{
				Field:     ops[0],
				Direction: ops[1],
			})
		}
	}

	rs, err := s.ws.GetMessageHistory(rq)
	if err != nil {
		s.ws.httpServer.respondWithError(writer, http.StatusInternalServerError, err.Message)
		return
	}

	s.ws.httpServer.respondWithJSON(writer, http.StatusOK, rs)

}
