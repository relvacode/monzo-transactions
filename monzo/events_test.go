package monzo

import "testing"

var txCreatedRaw = []byte(`{
    "type": "transaction.created",
    "data": {
        "account_id": "acc_00008gju41AHyfLUzBUk8A",
        "amount": -350,
        "created": "2015-09-04T14:28:40Z",
        "currency": "GBP",
        "description": "Ozone Coffee Roasters",
        "id": "tx_00008zjky19HyFLAzlUk7t",
        "category": "eating_out",
        "is_load": false,
        "settled": "2015-09-05T14:28:40Z",
        "merchant": {
            "address": {
                "address": "98 Southgate Road",
                "city": "London",
                "country": "GB",
                "latitude": 51.54151,
                "longitude": -0.08482400000002599,
                "postcode": "N1 3JD",
                "region": "Greater London"
            },
            "created": "2015-08-22T12:20:18Z",
            "group_id": "grp_00008zIcpbBOaAr7TTP3sv",
            "id": "merch_00008zIcpbAKe8shBxXUtl",
            "logo": "https://pbs.twimg.com/profile_images/527043602623389696/68_SgUWJ.jpeg",
            "emoji": "🍞",
            "name": "The De Beauvoir Deli Co.",
            "category": "eating_out"
        }
    }
}`)

func TestGetEvent(t *testing.T) {
	ev, err := GetEvent(txCreatedRaw)
	if err != nil {
		t.Fatal(err)
	}

	tc, ok := ev.(*TransactionCreated)
	if !ok {
		t.Fatal("expected a TransactionCreated type")
	}

	if tc.ID != "tx_00008zjky19HyFLAzlUk7t" {
		t.Fatalf("Bad decode: expected ID of %q but got %q", "tx_00008zjky19HyFLAzlUk7t", tc.ID)
	}
}

func BenchmarkGetEvent(b *testing.B) {
	for i := 0; i < b.N; i ++ {
		_, err := GetEvent(txCreatedRaw)
		if err != nil {
			b.Fatal(err)
		}
	}
}
