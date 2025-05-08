
export interface UrlData {
    path: string
    url: string
    owner: string
    modify: boolean
}

export interface QrData {
  path: string
  url: string
}

export interface ShortenedURL {
    path: string;
    url: string;
    owner: string
    modify: boolean
  }

export type ShortenedURLs = ShortenedURL[];
  