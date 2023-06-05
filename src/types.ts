import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface Membership {
  membershipType: number;
  membershipId: string;
  bungieName: string;
}

export interface MyQuery extends DataQuery {
  profile?: Membership;
  characters?: string[];
  activityMode?: number;
}

export const DEFAULT_QUERY: Partial<MyQuery> = {};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
}

export interface TrialsReportSearchResult {
  bnetId: number;
  bungieName: string;
  displayName: string;
  membershipId: string;
  membershipType: number;
  crossSaveOverride: CrossSaveOverride;
  emblemHash: number;
  lastPlayed: Date;
  score: number;
}

export interface CrossSaveOverride {
  membershipId: string;
  membershipType: number;
}

export interface CharacterItem {
  characterId: string;
  description: string;
  isPlaceholder?: boolean;
}
