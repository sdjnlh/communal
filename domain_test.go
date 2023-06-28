package communal

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInt64Array_MarshalJSON(t *testing.T) {
	arr := Int64Array{1, 2}
	bts, _ := arr.MarshalJSON()
	t.Log(string(bts))
}

func TestInt64Array_ToDB(t *testing.T) {
	arr := Int64Array{1, 2}

	bts, _ := arr.ToDB()
	t.Log(string(bts))
}

func TestInt64Array_FromDB(t *testing.T) {
	arr := Int64Array{}
	err := arr.FromDB([]byte("{1,2}"))

	if err != nil {
		t.Error(err)
	} else {
		t.Log(arr)
	}
}

type dateVO struct {
	Date string
}

func TestDate_UnmarshalJSON(t *testing.T) {
	var date UTCDate
	err := date.UnmarshalJSON([]byte("\"2020-03-07\""))

	if err != nil {
		t.Error(err)
	} else {
		t.Log(date)
		t.Log(time.Time(date).Format(time.RFC3339))
	}
	nilDateBs := []byte{}
	var nilDate UTCDate
	err = nilDate.UnmarshalJSON(nilDateBs)
	t.Log("nil date", err, nilDate)

	emptyDateBs := []byte("\"\"")
	t.Log("empty string length", len(emptyDateBs), string(emptyDateBs))
	var emptyDate UTCDate
	err = emptyDate.UnmarshalJSON(emptyDateBs)
	t.Log("empty date", err, nilDate)

	jsonStr := `{"date":""}`

	var d dateVO
	err = json.Unmarshal([]byte(jsonStr), &d)
	t.Log("date vo", d, err)
	err = emptyDate.UnmarshalJSON([]byte(d.Date))
	t.Log("empty date", err, nilDate)
}
