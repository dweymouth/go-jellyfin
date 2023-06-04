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
	PlayCount  int  `json:"PlayCount"`
	IsFavorite bool `json:"IsFavorite"`
	Played     bool `json:"Played"`
}

type NameID struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
}

type Images struct {
	Primary string `json:"Primary"`
}

type Song struct {
	Name           string   `json:"Name"`
	Id             string   `json:"Id"`
	Duration       int64    `json:"RunTimeTicks"`
	ProductionYear int      `json:"ProductionYear"`
	IndexNumber    int      `json:"IndexNumber"`
	Type           string   `json:"Type"`
	AlbumID        string   `json:"AlbumId"`
	Album          string   `json:"Album"`
	DiscNumber     int      `json:"ParentIndexNumber"`
	Artists        []NameID `json:"ArtistItems"`

	UserData UserData `json:"UserData"`
}

type songs struct {
	Songs      []*Song `json:"Items"`
	TotalSongs int     `json:"TotalRecordCount"`
}

type Artist struct {
	Name          string   `json:"Name"`
	ID            string   `json:"Id"`
	TotalDuration int64    `json:"RunTimeTicks"`
	Type          string   `json:"Type"`
	TotalSongs    int      `json:"SongCount"`
	TotalAlbums   int      `json:"AlbumCount"`
	UserData      UserData `json:"UserData"`
}

type artists struct {
	Artists      []*Artist `json:"Items"`
	TotalArtists int       `json:"TotalRecordCount"`
}

type Album struct {
	Name      string   `json:"Name"`
	ID        string   `json:"Id"`
	Duration  int64    `json:"RunTimeTicks"`
	Year      int      `json:"ProductionYear"`
	Type      string   `json:"Type"`
	Artists   []NameID `json:"AlbumArtists"`
	Overview  string   `json:"Overview"`
	Genres    []string `json:"Genres"`
	ImageTags Images   `json:"ImageTags"`
	UserData  UserData `json:"UserData"`
}

type albums struct {
	Albums      []*Album `json:"Items"`
	TotalAlbums int      `json:"TotalRecordCount"`
}

type Playlist struct {
	Name      string   `json:"Name"`
	ID        string   `json:"Id"`
	Genres    []string `json:"Genres"`
	Duration  int64    `json:"RunTimeTicks"`
	Type      string   `json:"Type"`
	MediaType string   `json:"MediaType"`
	SongCount int      `json:"ChildCount"`
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
