package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/modules/incubator/nft"
	autypes "github.com/cosmos/sdk-tutorials/auction/x/auction/internal/types"
)

type Keeper struct {
	BankKeeper bank.Keeper

	NftKeeper nft.Keeper

	StoreKey types.StoreKey

	cdc *codec.Codec
}

func NewKeeper(cdc *codec.Codec, bk bank.Keeper, nftk nft.Keeper, sKey types.StoreKey) Keeper {
	return Keeper{
		BankKeeper: bk,
		NftKeeper:  nftk,
		StoreKey:   sKey,
		cdc:        cdc,
	}
}

func (k Keeper) GetAuction(ctx types.Context, nftID string) (autypes.Auction, bool) {
	store := ctx.KVStore(k.StoreKey)
	nftIDByte := []byte(nftID)
	if !k.hasAuction(ctx, nftIDByte) {
		return autypes.Auction{}, false
	}
	bz := store.Get(nftIDByte)
	var auction autypes.Auction
	k.cdc.MustUnmarshalBinaryBare(bz, &auction)
	return auction, true
}

func (k Keeper) hasAuction(ctx types.Context, name []byte) bool {
	store := ctx.KVStore(k.StoreKey)
	return store.Has([]byte(name))
}

func (k Keeper) SetAuction(ctx types.Context, nftID string, auction autypes.Auction) {
	store := ctx.KVStore(k.StoreKey)
	nftIDByte := []byte(nftID)
	store.Set(nftIDByte, k.cdc.MustMarshalBinaryBare(auction))
}

func (k Keeper) DeleteAuction(ctx types.Context, nftID string) {
	store := ctx.KVStore(k.StoreKey)
	nftIDByte := []byte(nftID)
	store.Delete(nftIDByte)
}

// NewAuction creates a new auction for nfts,
func (k Keeper) NewAuction(ctx types.Context, nftID, nftDenom string, startTime, endTime time.Time) {
	auction := autypes.NewAuction(nftID, nftDenom, startTime, endTime)
	k.SetAuction(ctx, nftID, auction)
}

func (k Keeper) NewBid(ctx types.Context, nftID string, bidder types.AccAddress, bid types.Coins) {
	auction, ok := k.GetAuction(ctx, nftID)
	if !ok {
		return
	}
	// check the endtime
	if auction.EndTime.Before(ctx.BlockHeader().Time) {
		autypes.ErrAuctionOver(autypes.DefaultCodespace)
		return
	}
	newBid := autypes.NewBid(bidder, bid, nftID)
	auction.ReplaceBid(newBid)
	k.SetAuction(ctx, nftID, auction)
}

// Get an iterator over all names in which the keys are the names and the values are the whois
func (k Keeper) GetAuctionsIterator(ctx types.Context) types.Iterator {
	store := ctx.KVStore(k.StoreKey)
	return types.KVStorePrefixIterator(store, nil)
}

// func ()

// IterateAuctionsQueue iterates over the proposals in the inactive proposal queue
// and performs a callback function
func (keeper Keeper) IterateAuctionsQueue(ctx sdk.Context, cb func(auction autypes.Auction) (stop bool)) {
	iterator := keeper.GetAuctionsIterator(ctx)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		auctionId := string(iterator.Key())
		auction, found := keeper.GetAuction(ctx, auctionId)
		if !found {
			panic(fmt.Sprintf("proposal %d does not exist", auctionId))
		}

		if cb(auction) {
			break
		}
	}
}
