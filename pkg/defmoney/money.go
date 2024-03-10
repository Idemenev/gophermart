package defmoney

import (
	"bytes"
	"fmt"
	"github.com/Rhymond/go-money"
	"regexp"
	"strconv"
	"strings"
)

// пакет извращений ¯\_(ツ)_/¯
// работает только с положительными

const defaultCurrencyCode = "RUB"

func init() {
	money.UnmarshalJSON = defaultUnmarshalJSON
	// MarshalJSON is injection point of json.Marshaller for money.Money
	money.MarshalJSON = defaultMarshalJSON
}

func New(amount int64) *money.Money {
	return money.New(amount, defaultCurrencyCode)
}

func defaultUnmarshalJSON(m *money.Money, b []byte) error {
	data := string(b)

	amount, err := parseMoneyStrIntoInt(data, getFraction())
	if err != nil {
		return err
	}

	ref := money.New(amount, defaultCurrencyCode)

	*m = *ref
	return nil
}

func getFraction() int {
	return money.GetCurrency(defaultCurrencyCode).Fraction
}

func parseMoneyStrIntoInt(data string, fraction int) (int64, error) {
	// @todo this piece of shit assumes formats 12.10 or 12, not 12.1
	if data == "" {
		return 0, money.ErrInvalidJSONUnmarshal
	}

	re := fmt.Sprintf(`^(([1-9]\d*(.\d{%d})?)|0)$`, fraction)

	match, err := regexp.MatchString(re, data)
	if err != nil || !match {
		return 0, money.ErrInvalidJSONUnmarshal
	}
	if strings.Contains(data, ".") {
		data = strings.Replace(data, ".", "", 1)
	} else {
		data += strings.Repeat("0", fraction)
	}
	amount, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, money.ErrInvalidJSONUnmarshal
	}
	return amount, nil
}

func defaultMarshalJSON(m money.Money) ([]byte, error) {
	if m == (money.Money{}) {
		m = *money.New(0, "")
	}
	if m.Amount() == 0 {
		return bytes.NewBufferString("0").Bytes(), nil
	}
	moneyStr := strconv.FormatInt(m.Amount(), 10)

	fraction := getFraction()

	res := ""

	if fraction > len(moneyStr) {
		res = "0." + strings.Repeat("0", fraction-len(moneyStr)) + moneyStr
	} else {
		res = moneyStr[:len(moneyStr)-fraction] + "." + moneyStr[len(moneyStr)-fraction:]
	}

	return bytes.NewBufferString(res).Bytes(), nil
}
