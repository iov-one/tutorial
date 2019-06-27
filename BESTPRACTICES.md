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

        There are 2 critical points that needs to be explained in state models and messages.

        Firstly, You must have noticed one and said what the hell is

        ```protobuf
        weave.Metadata metadata = 1;
        ```

        `weave.Metadata` allows you to use *Migration* extension. `Migration extension` allows upgrading the code on chain without any downtime. This will be explained in more detail in further sections.

        Second one is how the magic `ID` field works. This will be explained in [Models](#Models) section.

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

    You must have notice we even validate if ```ID```'s lenght is not 0 and equal to 8 and tickers are actually string tickers. **Remember** more validation more solid your application is. If you **constrain** possible inputs, it is **less** checks you must do in the business logic.

3. #### Models
    > [PR#6](https://github.com/iov-one/tutorial/pull/6): _Create models_

    We defined our state in [codec section](#codec). In order to use models in weave we have to wrap our model with some functionalities and enforce it is a **morm.Model**

    Ensure our *OrderBook* fulfills **morm.Model**. This is just a helper so the compiler will complain loudly here if you forget to implement a method. Guaranteeing it *I am trying to implement this interface*.

    ```go
    var _ morm.Model = (*OrderBook)(nil)
    ```
    
    Now lets work on our models identity
    > How identity works in weave?

    - #### Auto incremented identities

        `morm.Model` covers auto incremented IDs for you. All you have to define `GetID` and `SetID` methods. If you defined `bytes id = 2 [(gogoproto.customname) = "ID"];`  on `codec.proto` you do not even need to write `GetID` method by yourself, Thanks to prototool it will be generated automatically. You will only need to define *SetID* method.

    - #### Custom identity

        For using your custom identity do not define `bytes id = 2 [(gogoproto.customname) = "ID"];` on `codec.proto`. You can use any other field as index with logic on top. Just write `SetID` and `GetID` logic that uses your index.

    Now you see how indexing works in **weave**. Lets see the other wonders of **weave**.

    In order to our model to fulfill **Model** it must be [Clonable](https://github.com/iov-one/weave/blob/master/orm/interfaces.go#L34).

    This is how you ensure it:

    ```go
    func (o *OrderBook) Copy() orm.CloneableData {
	    return &OrderBook{
		    Metadata:      o.Metadata.Copy(),
		    ID:            copyBytes(o.ID),
		    MarketID:      copyBytes(o.MarketID),
		    AskTicker:     o.AskTicker,
		    BidTicker:     o.BidTicker,
		    TotalAskCount: o.TotalAskCount,
		    TotalBidCount: o.TotalBidCount,
	    }
    }
    ```

    Another point **Model** enforces is ```Validate``` method.

    ```go
    func (o *OrderBook) Validate() error {
	    if err := isGenID(o.ID, true); err != nil {
    		return err
	    }
	    if err := isGenID(o.MarketID, false); err != nil {
		    return errors.Wrap(err, "market id")
	    }
	    if !coin.IsCC(o.AskTicker) {
		    return errors.Wrap(errors.ErrModel, "invalid ask ticker")
	    }
	    if !coin.IsCC(o.BidTicker) {
		    return errors.Wrap(errors.ErrModel, "invalid bid ticker")
	    }
	    if o.TotalAskCount < 0 {
		    return errors.Wrap(errors.ErrModel, "negative total ask count")
	    }
	    if o.TotalBidCount < 0 {
    		return errors.Wrap(errors.ErrModel, "negative total bid count")
    	}
	    return nil
    }
    ```

    >I want to point out again and make it persistent in your mind: Extensive **Validation** is crucial.

4. #### Buckets
    > [PR#8](https://github.com/iov-one/tutorial/pull/8): _Create buckets_
    \
    [PR#9](https://github.com/iov-one/tutorial/pull/9): _Add indexer for market using marketID, askTicker, bidTicker_
    
    Buckets are the components that we will use to interact with the KV store. It is our data warehouse. 
    \
    Check out [morm](https://github.com/iov-one/tutorial/blob/master/morm/model_bucket.go#L40) package. It is enhanced version of [weave/orm](https://github.com/iov-one/weave/tree/master/orm) with indexes for making queries easier. 
    >***FYI*** whenever you have questions in your mind about the internals check out ```weave``` source code. It is very well documented. Please feel free to add [issues](https://github.com/iov-one/weave/issues) if you think there is something under documented or confusing. 
    
    Lets dive in to code now.

    First define ```MarketBucket``` that will hold ```Market``` informations and write a function creates a market. This is a basic ```morm/model_bucket``` without any indexes. 

    ```go
    type MarketBucket struct {
        morm.ModelBucket
    }

    func NewMarketBucket() *MarketBucket {
	    b := morm.NewModelBucket("market", &Market{})
	    return &MarketBucket{
		    ModelBucket: b,
	    }
    }
    ```

    Now create your ```OrderBookBucket``` with indexes. Indexes enable us inserting and querying models with ease. You can think is as SQL indexes.

    ```go
    type OrderBookBucket struct {
    	morm.ModelBucket
    }

    // NewOrderBookBucket initates orderbook with required indexes
    func NewOrderBookBucket() *OrderBookBucket {
	    b := morm.NewModelBucket("orderbook", &OrderBook{},
		    morm.WithIndex("market", marketIDindexer, false),
		    morm.WithIndex("marketWithTickers", marketIDTickersIndexer, true),
        )
    
	    return &OrderBookBucket{
	    	ModelBucket: b,
        }
    }
    ```

    *marketIDindexer* is an index with only market id. This is a simple index form to implement.
    It just bindinds OrderBook to a MarketId(*bytes*)

    In ```weave``` we use uniformed *bytes* as indexes. This improves performance.

    ```go
    func marketIDindexer(obj orm.Object) ([]byte, error) {
    	if obj == nil || obj.Value() == nil {
	    	return nil, nil
	    }
	    ob, ok := obj.Value().(*OrderBook)
	    if !ok {
		    return nil, errors.Wrapf(errors.ErrState, "expected orderbook, got %T", obj.Value())
	    }
	    return ob.MarketID, nil
    }
    ```
    
    You must have ideas flying around on your mind like **how are we going to make an compound index? Really!? Is it all weave has?**

    \
    Do not worry. **Weave** is like a swiss knife with a lot of blockchain features.

    Now it is time for a compound index.
    Here is how we create compound index for morm buckets:

    ```go
    func BuildMarketIDTickersIndex(orderbook *OrderBook) []byte {
	    askTickerByte := make([]byte, tickerByteSize)
	    copy(askTickerByte, orderbook.AskTicker)

	    bidTickerByte := make([]byte, tickerByteSize)
	    copy(bidTickerByte, orderbook.BidTicker)

        return bytes.Join([][]byte{orderbook.MarketID, askTickerByte, bidTickerByte}, nil)
    }
    ```
    
    *BuildMarketIDTickersIndex =  indexByteSize = 8(MarketID) + ask ticker size + bid ticker size*

    Sample market id index with tickers = 000000056665820070797900

    000000056665820070797900 = 00000005(**MarketID = 5**) + 6665200(**BAR ticker in bytes**) + 70797900(**FOO ticker in bytes**)

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