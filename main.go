package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/trie"
)

type CacheConfig struct {
	TrieCleanLimit      int           // Memory allowance (MB) to use for caching trie nodes in memory
	TrieCleanJournal    string        // Disk journal for saving clean cache entries.
	TrieCleanRejournal  time.Duration // Time interval to dump clean cache to disk periodically
	TrieCleanNoPrefetch bool          // Whether to disable heuristic state prefetching for followup blocks
	TrieDirtyLimit      int           // Memory limit (MB) at which to start flushing dirty trie nodes to disk
	TrieDirtyDisabled   bool          // Whether to disable trie write caching and GC altogether (archive node)
	TrieTimeLimit       time.Duration // Time limit after which to flush the current in-memory trie to disk
	SnapshotLimit       int           // Memory allowance (MB) to use for caching snapshot entries in memory
	Preimages           bool          // Whether to store preimage of trie key to the disk

	SnapshotWait bool // Wait for snapshot construction on startup. TODO(karalabe): This is a dirty hack for testing, nuke it
}

// defaultCacheConfig are the default caching values if none are specified by the
// user (also used during testing).
var defaultCacheConfig = &CacheConfig{
	TrieCleanLimit:   16 * 1024,
	TrieDirtyLimit:   16 * 1024,
	TrieTimeLimit:    1 * time.Second,
	TrieCleanJournal: "triejournal",
	SnapshotLimit:    256,
	SnapshotWait:     true,
}
var CacheSize = common.StorageSize(defaultCacheConfig.TrieDirtyLimit * 1024 * 1024)

func test_leveldb(dbpath string) {
	keystr := "df3f619804a92fdb4057192dc43dd748ea778adc52bc498ce80524c014b81119"
	key, _ := hex.DecodeString(keystr)
	level, err := rawdb.NewLevelDBDatabase(dbpath, 1024*2, 1024, "", false)
	if err != nil {
		fmt.Println(err)
		return
	}
	level.Put(key, key)
	level.Close()

	level2, err := rawdb.NewLevelDBDatabase(dbpath, 1024*2, 1024, "", false)
	if err != nil {
		fmt.Println(err)
		return
	}
	val, err := level2.Get(key)
	if err != nil {
		fmt.Println(err)
		return
	}
	valstr := hex.EncodeToString(val)
	fmt.Println(valstr)
}

func test_trie(dbpath string) {
	keystr := "df3f619804a92fdb4057192dc43dd748ea778adc52bc498ce80524c014b81119"
	key, _ := hex.DecodeString(keystr)

	level, err := rawdb.NewLevelDBDatabase(dbpath, 1024*2, 1024, "", false)
	if err != nil {
		fmt.Println(err)
		return
	}
	stateCache := state.NewDatabaseWithConfig(level, &trie.Config{
		Cache:     defaultCacheConfig.TrieCleanLimit,
		Journal:   defaultCacheConfig.TrieCleanJournal,
		Preimages: defaultCacheConfig.Preimages,
	})

	trie, err := stateCache.OpenTrie(common.Hash{})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = trie.TryUpdate(key, key)
	if err != nil {
		fmt.Println(err)
		return
	}

	val, err := trie.TryGet(key)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hex.EncodeToString(val))

	trie.Commit(nil)
	stateCache.TrieDB().Commit(trie.Hash(), false, nil)
	fmt.Println(trie.Hash())

	rhash := common.HexToHash("1441908d7e391c5db3760dc4a1b3008169f629053fc2c061bd436bc19543aba5")
	trie2, err := stateCache.OpenTrie(rhash)
	if err != nil {
		fmt.Println(err)
		return
	}

	val, _ = trie2.TryGet(key)
	fmt.Println(hex.EncodeToString(val))
}

func test_subtrie(dbpath string) {
	level, err := rawdb.NewLevelDBDatabase(dbpath, 1024*2, 1024, "", false)
	if err != nil {
		fmt.Println(err)
		return
	}
	stateCache := state.NewDatabaseWithConfig(level, &trie.Config{
		Cache:     defaultCacheConfig.TrieCleanLimit,
		Journal:   defaultCacheConfig.TrieCleanJournal,
		Preimages: defaultCacheConfig.Preimages,
	})

	_, err = stateCache.OpenTrie(common.Hash{})
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = stateCache.OpenTrie(common.HexToHash("d7f8974fb5ac78d9ac099b9ad5018bedc2ce0a72dad1827a1709da30580f0544"))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	test_subtrie("testdb")
}
