package accrual

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

type DefaultAccrual struct {
	client http.Client
	addr   string
}

func NewDefaultAccrualService(addr string) Service {
	client := http.Client{}
	client.Timeout = time.Second * 5

	return &DefaultAccrual{
		client: client,
		addr:   addr,
	}
}

func (s DefaultAccrual) GetOrderInfo(orderID string) (OrderInfo, error) {
	var info OrderInfo
	response, err := s.client.Get(s.addr + "/api/orders/" + orderID)
	if err != nil {
		return info, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return info, errors.New("bad status code: " + strconv.Itoa(response.StatusCode))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}
