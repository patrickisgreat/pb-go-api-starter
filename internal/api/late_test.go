package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/patrickisgreat/pb-go-api-starter/internal/appserver"
	proto "github.com/patrickisgreat/pb-go-api-starter/internal/generated/pb_go_api_starter/api"
	"github.com/patrickisgreat/pb-go-api-starter/internal/generated/pb_go_api_starter/session"
	"github.com/patrickisgreat/pb-go-api-starter/internal/quoterepo"
	"github.com/patrickisgreat/pb-go-api-starter/internal/quoteservice"
)

func TestGetQuoteLateHandler(t *testing.T) {
	c := qt.New(t)

	// Build a QuoteService backed by a single known quote
	quoteService := quoteservice.NewQuoteService(&quoterepo.QuoteRepository{
		Quotes: []string{"foo"},
	})

	// Create a mock ServerContext
	serverContext := &appserver.ServerContext{
		QuoteService: quoteService,
	}

	// Create a request
	request := &proto.GetQuoteLateRequest{
		UserSession: &session.UserSession{
			UserUrn:  &wrapperspb.StringValue{Value: "myapp:users:2"},
			Features: []string{},
		},
		DelayMs: &wrapperspb.Int32Value{Value: 500},
	}

	// Capture the start time
	s := time.Now()

	// Call the handler
	response, err := GetQuoteLateHandler(context.Background(), serverContext, request)

	// Assert the results
	c.Assert(err, qt.IsNil)
	c.Assert(response, qt.Not(qt.IsNil))
	c.Assert(response.Quote, qt.Equals, "foo")

	f := time.Since(s)
	if f.Milliseconds() <= int64(500) {
		fmt.Printf("f.Milliseconds() = %d\n", f.Milliseconds())
	}
	c.Assert(f.Milliseconds() >= int64(500), qt.IsTrue)
}

func TestGetQuoteLateHandler_InvalidSession(t *testing.T) {
	c := qt.New(t)

	// Build a QuoteService backed by a single known quote
	quoteService := quoteservice.NewQuoteService(&quoterepo.QuoteRepository{
		Quotes: []string{"foo"},
	})

	// Create a mock ServerContext
	serverContext := &appserver.ServerContext{
		QuoteService: quoteService,
	}

	// Create a request
	request := &proto.GetQuoteLateRequest{
		UserSession: &session.UserSession{
			UserUrn:  &wrapperspb.StringValue{Value: "myapp:tracks:2"},
			Features: []string{},
		},
	}

	// Call the handler
	_, err := GetQuoteLateHandler(context.Background(), serverContext, request)

	// Assert the results
	c.Assert(err, qt.IsNotNil)
	c.Assert(err.Error(), qt.Equals, "twirp error invalid_argument: userSession.userUrn invalid user urn")
}
