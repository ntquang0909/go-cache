package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChain(t *testing.T) {
	var chain Cache = NewChain(
		NewMongoDBStore(MongoDBStoreOptions{
			DatabaseURI:  "mongodb://localhost:27012",
			DatabaseName: "test_cache",
			Entity:       "caches",
		}),
		NewMemoryStore(MemoryStoreOptions{}),
		NewRistrettoStore(RistrettoStoreOptionsDefault),
		NewMemcacheStore(&MemcacheStoreOptions{
			Servers: []string{"localhost:11211"},
		}),
		NewRedisStore(&RedisStoreOptions{
			Address: "localhost:6379",
		}),
	)
	var err error

	// Test string
	var strKey = "test_str_key"
	var strIn = "Hello world"
	err = chain.Set(strKey, &strIn)
	assert.NoError(t, err)

	var strOut string
	err = chain.Get(strKey, &strOut)
	assert.NoError(t, err)

	assert.Equal(t, strOut, strIn)

	// Test struct
	var itemIn = AutoGenerated{
		ID:          "jt_standard_small_c0cbv999vbk2btpr6b70",
		RateID:      "jt_standard_small",
		TotalFee:    "3.21",
		ShippingFee: "3.21",
		Tax:         "0",
		DeliveryService: DeliveryService{
			ID:                    "jt_standard",
			Code:                  "DOM123",
			Name:                  "Standard Delivery (1-3 days)",
			Description:           "Standard Delivery (1-3 days)",
			Courier:               "jt",
			EstimatedDeliveryTime: "1 - 3 Working Days",
		},
		DeliveryServiceDescription: "Next 3 days delivery",
		Courier: &Courier{
			Alias:         "jt",
			Name:          "J\u0026T",
			ImageURL:      "https://ezielog-staging.s3-ap-southeast-1.amazonaws.com/icons/courier/j\u0026t.svg",
			Rating:        5,
			Tracking:      "",
			TermURL:       "https://www.jtexpress.my/tnc.php",
			PrivacyURL:    "https://www.jtexpress.my/tnc.php",
			ContactNumber: "(+65) 6939 6399",
			SupportURL:    "https://www.jtexpress.sg/contact-us",
		},
		Rating:       5,
		Currency:     "SGD",
		ProductType:  "",
		DeliveryType: "pick_up",
		Dimension: Dimension{
			Weight: 1,
			Length: 38,
			Width:  15,
			Height: 20,
		},
		PackageType: "handbag",
	}
	var itemKey = "test_item_key"
	err = chain.Set(itemKey, &itemIn)
	assert.NoError(t, err)

	var itemOut AutoGenerated
	err = chain.Get(itemKey, &itemOut)
	assert.NoError(t, err)

	assert.Equal(t, itemOut, itemIn)

	// Test bool
	var boolKey = "test_bool_key"
	var boolIn = true
	err = chain.Set(boolKey, &boolIn, time.Hour)
	assert.NoError(t, err)

	var boolOut bool
	err = chain.Get(boolKey, &boolOut)
	assert.NoError(t, err)

	assert.Equal(t, boolIn, boolOut)
}
