export interface SonnarQueueResponse {
    Item:        Item;
    LastChecked: Date;
    FirstSeen: Date;
}

export interface Item {
    series:                  Series;
    episode:                 Episode;
    quality:                 ItemQuality;
    size:                    number;
    title:                   string;
    sizeleft:                number;
    timeleft:                string;
    estimatedCompletionTime: Date;
    status:                  string;
    trackedDownloadStatus:   string;
    statusMessages:          any[];
    downloadId:              string;
    protocol:                string;
    id:                      number;
}

export interface Episode {
    seriesId:                 number;
    episodeFileId:            number;
    seasonNumber:             number;
    episodeNumber:            number;
    title:                    string;
    airDate:                  Date;
    airDateUtc:               Date;
    overview:                 string;
    hasFile:                  boolean;
    monitored:                boolean;
    absoluteEpisodeNumber:    number;
    unverifiedSceneNumbering: boolean;
    lastSearchTime:           Date;
    id:                       number;
}

export interface ItemQuality {
    quality:  QualityQuality;
    revision: Revision;
}

export interface QualityQuality {
    id:         number;
    name:       string;
    source:     string;
    resolution: number;
}

export interface Revision {
    version:  number;
    real:     number;
    isRepack: boolean;
}

export interface Series {
    title:             string;
    sortTitle:         string;
    seasonCount:       number;
    status:            string;
    overview:          string;
    network:           string;
    airTime:           string;
    images:            Image[];
    seasons:           Season[];
    year:              number;
    path:              string;
    profileId:         number;
    languageProfileId: number;
    seasonFolder:      boolean;
    monitored:         boolean;
    useSceneNumbering: boolean;
    runtime:           number;
    tvdbId:            number;
    tvRageId:          number;
    tvMazeId:          number;
    firstAired:        Date;
    lastInfoSync:      Date;
    seriesType:        string;
    cleanTitle:        string;
    imdbId:            string;
    titleSlug:         string;
    certification:     string;
    genres:            string[];
    tags:              any[];
    added:             Date;
    ratings:           Ratings;
    qualityProfileId:  number;
    id:                number;
}

export interface Image {
    coverType: string;
    url:       string;
}

export interface Ratings {
    votes: number;
    value: number;
}

export interface Season {
    seasonNumber: number;
    monitored:    boolean;
}
