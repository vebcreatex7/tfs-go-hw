package handlers

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tfs-go-hw/course_project/internal/domain"
)

type BotServiceMock struct{}

func (bsm *BotServiceMock) Run(stop <-chan struct{}, stoped chan<- struct{}) {
	<-stop
	stoped <- struct{}{}
}

func (bsm *BotServiceMock) SetSymbol(string) {}

func (bsm *BotServiceMock) GetSymbol() string {
	return "pi_xbtusd"
}

func (bsm *BotServiceMock) SetPeriod(domain.CandlePeriod) {}

func (bsm *BotServiceMock) GetPeriod() domain.CandlePeriod {
	return domain.CandlePeriod1m
}

func (bsm *BotServiceMock) SetAmount(amount int) {}

func (bsm *BotServiceMock) ChangeSourceIndicator(s rune) {}

func (bsm *BotServiceMock) ConfigurateIndicator(fast, slow, signal int, s rune) {}

func TestHandlers_Testify(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	botServiceMock := &BotServiceMock{}
	bot := NewBot(ctx, botServiceMock, nil)
	bot.isRunning = false

	// Exchange
	body := `{"symbol":"pi_xbtusd","period":"1m","amount":10}`
	req := httptest.NewRequest(http.MethodPost, "/exchange/config", strings.NewReader(body))
	w := httptest.NewRecorder()
	bot.ConfigurateExchange(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	expect := "Ok"
	assert.Equal(t, expect, string(data))

	// Indicator
	body = `{"fast":10,"slow":25,"signal":9, "source":"c"}`
	req = httptest.NewRequest(http.MethodPost, "/indicator/config", strings.NewReader(body))
	w = httptest.NewRecorder()
	bot.ConfigurateIndicator(w, req)
	res = w.Result()
	defer res.Body.Close()
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	expect = "Ok"
	assert.Equal(t, expect, string(data))

	// Source
	body = `{"source":"o"}`
	req = httptest.NewRequest(http.MethodPost, "/indicator/change_source", strings.NewReader(body))
	w = httptest.NewRecorder()
	bot.ChangeSourceIndicator(w, req)
	res = w.Result()
	defer res.Body.Close()
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	expect = "Ok"
	assert.Equal(t, expect, string(data))

}
