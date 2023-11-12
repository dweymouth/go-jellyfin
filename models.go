package jellyfin

type ItemType string

const (
	TypeArtist   ItemType = "Artist"
	TypeAlbum    ItemType = "Album"
	TypePlaylist ItemType = "Playlist"
	//	TypeQueue    ItemType = "Queue"
	//	TypeHistory  ItemType = "History"
	TypeSong  ItemType = "Song"
	TypeGenre ItemType = "Genre"
)

type UserData struct {
	PlayCount      int    `json:"PlayCount"`
	IsFavorite     bool   `json:"IsFavorite"`
	Rating         int    `json:"Rating"`
	Played         bool   `json:"Played"`
	LastPlayedDate string `json:"LastPlayedDate"`
}

type NameID struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
}

type Images struct {
	Primary string `json:"Primary"`
	Disc    string `json:"Disc"`
}

type MediaSource struct {
	Bitrate   int    `json:"Bitrate"`
	Container string `json:"Container"`
	Path      string `json:"Path"`
	Size      int    `json:"Size"`
}

type Song struct {
	Name           string        `json:"Name"`
	Id             string        `json:"Id"`
	RunTimeTicks   int64         `json:"RunTimeTicks"`
	ProductionYear int           `json:"ProductionYear"`
	DateCreated    string        `json:"DateCreated"`
	IndexNumber    int           `json:"IndexNumber"`
	Type           string        `json:"Type"`
	AlbumID        string        `json:"AlbumId"`
	Album          string        `json:"Album"`
	DiscNumber     int           `json:"ParentIndexNumber"`
	Artists        []NameID      `json:"ArtistItems"`
	ImageTags      Images        `json:"ImageTags"`
	MediaSources   []MediaSource `json:"MediaSources"`
	UserData       UserData      `json:"UserData"`
}

type songs struct {
	Songs      []*Song `json:"Items"`
	TotalSongs int     `json:"TotalRecordCount"`
}

type Artist struct {
	Name         string   `json:"Name"`
	Overview     string   `json:"Overview"`
	ID           string   `json:"Id"`
	RunTimeTicks int64    `json:"RunTimeTicks"`
	Type         string   `json:"Type"`
	SongCount    int      `json:"SongCount"`
	AlbumCount   int      `json:"AlbumCount"`
	UserData     UserData `json:"UserData"`
	ImageTags    Images   `json:"ImageTags"`
}

type artists struct {
	Artists      []*Artist `json:"Items"`
	TotalArtists int       `json:"TotalRecordCount"`
}

type Album struct {
	Name         string   `json:"Name"`
	ID           string   `json:"Id"`
	RunTimeTicks int64    `json:"RunTimeTicks"`
	Year         int      `json:"ProductionYear"`
	DateCreated  string   `json:"DateCreated"`
	Type         string   `json:"Type"`
	Artists      []NameID `json:"AlbumArtists"`
	Overview     string   `json:"Overview"`
	Genres       []string `json:"Genres"`
	ChildCount   int      `json:"ChildCount"`
	ImageTags    Images   `json:"ImageTags"`
	UserData     UserData `json:"UserData"`
}

type albums struct {
	Albums      []*Album `json:"Items"`
	TotalAlbums int      `json:"TotalRecordCount"`
}

type Playlist struct {
	Name               string   `json:"Name"`
	ID                 string   `json:"Id"`
	Overview           string   `json:"Overview"`
	DateCreated        string   `json:"DateCreated"`
	DateLastMediaAdded string   `json:"DateLastMediaAdded"`
	Genres             []string `json:"Genres"`
	RunTimeTicks       int64    `json:"RunTimeTicks"`
	Type               string   `json:"Type"`
	MediaType          string   `json:"MediaType"`
	ImageTags          Images   `json:"ImageTags"`
	SongCount          int      `json:"ChildCount"`
}

type playlists struct {
	Playlists      []*Playlist `json:"Items"`
	TotalPlaylists int         `json:"TotalRecordCount"`
}

type SearchResult struct {
	Artists   []*Artist
	Albums    []*Album
	Songs     []*Song
	Playlists []*Playlist
}
