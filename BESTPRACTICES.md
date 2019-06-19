# Best Practices
---

First of all to understand weave design philosophy and for the sake of programming well designed software, I advise you to read this fine-grained article:
[Things I Learnt The Hard Way (in 30 Years of Software Development)](https://blog.juliobiason.net/thoughts/things-i-learnt-the-hard-way/) \
\
This documents content is curated from PR discussions of weave tutorial. If you follow the *PRs* section you can see how to implement blockchain app using weave.
And you can see how the development flow works and what are the issues you should have on your mind while moving forward.

### Development flow
1. #### Codec
    > [PR#1](https://github.com/iov-one/tutorial/pull/1): _Create order book models_

    Codec is the first the component that needs to be designed. Keep in mind that this part is the most important since whole module will depend on *codec*. You can think codec it as *model* in mvc pattern. Yet it is not simple as model. Codec defines the whole application state models and more.

    - #### State
        ```protobuf
        message Trade {
            weave.Metadata metadata = 1;
            bytes id = 2 [(gogoproto.customname) = "ID"];
            bytes order_book_id = 3 [(gogoproto.customname) = "OrderBookID"];
            bytes order_id = 4 [(gogoproto.customname) = "OrderID"];
            bytes taker = 5 [(gogoproto.casttype) = "github.com/iov-one/weave.Address"];
            bytes maker = 6 [(gogoproto.casttype) = "github.com/iov-one/weave.Address"];
            coin.Coin maker_paid = 7;
            coin.Coin taker_paid = 8;
            int64 executed_at = 9 [(gogoproto.casttype) = "github.com/iov-one/weave.UnixTime"];
        }
        ```

        As you can see above, weave is heavily using `bytes` for identites, addresses etc. At first glance this might seem difficult to handle but if you take a look at [x/orderbook/bucket](https://github.com/iov-one/tutorial/blob/master/x/orderbook/bucket.go#L125) you can see it is very useful for indexing and performance

    - #### Message definitions
        ```protobuf
        message CreateOrderMsg {
            weave.Metadata metadata = 1;
            bytes trader = 2 [(gogoproto.casttype) = "github.com/iov-one/weave.Address"];
            bytes order_book_id = 3 [(gogoproto.customname) = "OrderBookID"];
            coin.Coin offer = 4;
            Amount price = 5;
        }
        ```

        As you can see above w*e define type necessary fields that will be used by `handler` to interact with application state

    After defining your state and messages run `make protoc` and make sure `prototool` creates your `codec.pb.go` file successfully.
    Now we have scaffold of our application thanks to auto-generated `codec.pb.go` file.

2. #### Messages
    > [PR#2](https://github.com/iov-one/tutorial/pull/2): _Create msgs_
    
    This is where we wrap our auto generated message types with useful and vital functionalities. Before getting into this section I want to remind you *validation* is **EXTREMELY**  important. First create your `msg.go` file. This is where the magic will happen :grin:

    Lets bind a path to our new message

    ```go
    const (
	    pathCreateOrderBook = "order/create_book"
    )
    ```

    After that ensure your message is a `weave.Msg`

    ```go
    var _ weave.Msg = (*CreateOrderBookMsg)(nil)
    ```

    ```go
    func (m CreateOrderBookMsg) Validate() error {
	    if err := m.Metadata.Validate(); err != nil {
		    return errors.Wrap(err, "metadata")
	    }
	    if err := validID(m.MarketID); err != nil {
		    return errors.Wrap(err, "market id")
	    }
	    if !coin.IsCC(m.AskTicker) {
		    return errors.Wrapf(errors.ErrCurrency, "Invalid Ask Ticker: %s", m.AskTicker)
	    }
	    if !coin.IsCC(m.BidTicker) {
		    return errors.Wrapf(errors.ErrCurrency, "Invalid Bid Ticker: %s", m.BidTicker)
	    }
	    if m.BidTicker <= m.AskTicker {
		    return errors.Wrapf(errors.ErrCurrency, "ask (%s) must be before bid (%s)", m.AskTicker, m.BidTicker)
	    }
	    return nil
    }
    ```

    ```go
    func validID(id []byte) error {
	    if len(id) == 0 {
		    return errors.Wrap(errors.ErrEmpty, "id missing")
	    }
	    if len(id) != 8 {
		    return errors.Wrap(errors.ErrInput, "id is invalid length (expect 8 bytes)")
	    }
	    return nil
    }
    ```

    You must have notice we even validate if ```ID```'s lenght is not 0 and equal to 8 and tickers are actually string tickers. **Remember** more validation more solid your application is.

## PRs
 - [PR#1](https://github.com/iov-one/tutorial/pull/1): _Create order book models_
 - [PR#2](https://github.com/iov-one/tutorial/pull/2): _Create msgs_
 - [PR#6](https://github.com/iov-one/tutorial/pull/6): _Create models_
 - [PR#8](https://github.com/iov-one/tutorial/pull/8): _Create buckets_
 - [PR#9](https://github.com/iov-one/tutorial/pull/9): _Add indexer for market using marketID, askTicker, bidTicker_
 - [PR#11](https://github.com/iov-one/tutorial/pull/10): _Create handlers_

## Discussions compiled from PRs
> Is weave.Metadata metadata always required?

This allows you to upgrade the code on chain without downtime
reference: [weave/migration](https://github.com/iov-one/weave/tree/master/migration/doc.go)

> What is the best practices for using proto value types on modeling currency amount, IDs, timestamp?  

 - Currency example [weave/x/coin](https://github.com/iov-one/weave/tree/master/coin)
 - For incremetal ID use `orm.IDGenBucket` _TODO: Add ```morm/ModelBucket``` when it is merged into weave_
 - For Timestamp [weave.UnixTime](https://github.com/iov-one/weave/blob/master/time.go)
  
> How to identify users in state.proto implementation

 - Use [weave.Address](https://github.com/iov-one/weave/blob/master/x/aswap/codec.proto)

### !! **Write document and specs for everything. Literally *EVERYTHING*** eg. [weave/tutorial/orderbook/codec](https://github.com/iov-one/tutorial/blob/master/x/orderbook/codec.proto) 

> Can you elaborate what m.Metadata.Validate() does in detail?
 - _[Discussion link](https://github.com/iov-one/tutorial/pull/2#discussion_r289601156)_
 - Currently, it just Validates that Metadata is not nil and there is a positive integer schema set. But in general, all structs used in public APIs have a Validate() error method. And when we embed one in another object, we can just Validate all the sub-structures and return the wrapped error, just giving the context why we were calling it (this makes more sense for a Coin which can be used in multiple fields, than a Metadata) 

> weave/migration NoModification and RefuseMigration difference
 - _[Discussion link](https://github.com/iov-one/tutorial/pull/2#discussion_r289602376)_
 - For registering the schema 1, you should always have migration.NoModification to initialize it. For future schemas, they can be a soft change (with a migration path), or a hard break (no migration path). Hard schema changes in models may be next to impossible to implement on a chain without a strong manual hard-fork.