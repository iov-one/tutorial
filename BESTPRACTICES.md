# Best Practices
---

This documents content is curated by the PR discussions of weave tutorial. If you follow the *PRs* section you can see how implement blockchain app using weave.
And you can see how the development flow works and what are the issues you should have on your mind while moving forward.

## PRs
 - [PR#1](https://github.com/iov-one/tutorial/pull/1): _Create order book models_
 - [PR#2](https://github.com/iov-one/tutorial/pull/2): _Create msgs_
 - [PR#6](https://github.com/iov-one/tutorial/pull/6): _Create models_
 - [PR#8](https://github.com/iov-one/tutorial/pull/8): _Create buckets_
 - [PR#9](https://github.com/iov-one/tutorial/pull/9): _Add indexer for market using marketID, askTicker, bidTicker_
 - [PR#11](https://github.com/iov-one/tutorial/pull/10): _Create handlers_
## Weave 
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