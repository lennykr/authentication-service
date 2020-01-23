package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"log"
)

// TODO: should use some discovery mechanism
const databaseUrl = "51.105.216.241:9080"

func newClient() *dgo.Dgraph {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial(databaseUrl, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}

func GetPasswordByEmail(ctx context.Context, email string) (string, error) {
	c := newClient()

	variables := map[string]string{"$email": email}
	q := `
		query x($email: string){
			email(func: eq(emailAddress, $email)) {
				emailAddress
				~email {
					password
				}
			}
		}
	`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	// defer txn.Discard(ctx)
	// resp, err := txn.Query(context.Background(), q)
	if err != nil {
		log.Fatal(err)
	}

	// After we get the balances, we have to decode them into structs so that
	// we can manipulate the data.
	var decode struct {
		All []struct {
			Address string `json:"emailAddress"`
			User    []struct {
				Password string `json:"password"`
			} `json:"~email"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return "", err
	}

	if len(decode.All) == 0 {
		return "", errors.New("couldn't find email")
	}

	return decode.All[0].User[0].Password, nil
}
