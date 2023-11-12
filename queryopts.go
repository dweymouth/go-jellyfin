package jellyfin

import "time"

type SortField string

const (
	SortByName            SortField = "SortName"
	SortByYear            SortField = "ProductionYear,PremiereDate"
	SortByArtist          SortField = "AlbumArtist"
	SortByPlayCount       SortField = "PlayCount"
	SortByRandom          SortField = "Random"
	SortByDateCreated     SortField = "DateCreated"
	SortByDatePlayed      SortField = "DatePlayed"
	SortByCommunityRating SortField = "CommunityRating"
)

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

// Sort describes sorting
type Sort struct {
	Field SortField
	Mode  SortOrder
}

type Paging struct {
	StartIndex int
	Limit      int
}

type FilterPlayStatus string

const (
	FilterIsPlayed    = "Played"
	FilterIsNotPlayed = "Not played"
)

// Filter contains filter for reducing results. Some fields are exclusive,
type Filter struct {
	// Played
	FilterPlayed FilterPlayStatus
	// Favorite marks items as being starred / favorite.
	Favorite bool
	// Include only results from the given artist.
	ArtistID string
	// Include only results with the given parentID.
	ParentID string
	// Genres contains list of genres to include.
	Genres []NameID
	// YearRange contains two elements, items must be within these boundaries.
	YearRange [2]int
}

type QueryOpts struct {
	Paging Paging
	Filter Filter
	Sort   Sort
}

func (f Filter) yearRangeValid() bool {
	if f.YearRange == [2]int{0, 0} {
		return true
	}

	if f.YearRange[0] > f.YearRange[1] {
		return false
	}

	if f.YearRange[0] < 0 {
		return false
	}

	year := time.Now().Year()
	return f.YearRange[1] <= year+10
}
