// A sponsor card as GET /api/sponsors returns it. The sponsor's own link is not in it: the card
// links to clickUrl, which counts the click and redirects.
export interface Sponsor {
  id: number;
  name: string;
  text: string;
  // empty if the sponsor has no picture
  imageUrl: string;
  clickUrl: string;
}

// A sponsor as the admin list shows it.
export interface AdminSponsor {
  id: number;
  name: string;
  text: string;
  link: string;
  // 2026-10-22, empty for no limit
  startsOn: string;
  endsOn: string;
  enabled: boolean;
  imageUrl: string;
  // the card is shown today
  active: boolean;
  impressions: number;
  clicks: number;
  impressions30: number;
  clicks30: number;
}

// What the admin sends to create or change a sponsor.
export interface SponsorForm {
  name: string;
  text: string;
  link: string;
  startsOn: string;
  endsOn: string;
  enabled: boolean;
  // a new picture as base64, left out to keep the sponsor's picture
  image?: string;
  removeImage?: boolean;
}
