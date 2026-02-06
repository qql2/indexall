export const APP_NAME = "IndexAll";

export type ResourceStatus = 'active' | 'stale' | 'deleted';

export interface Tag {
  id: string;
  name: string;
  color: string | null;
  aliases: string[];
  parentIds: string[];
}

export interface Resource {
  id: string;
  source: string;
  externalId: string | null;
  title: string;
  description: string | null;
  url: string | null;
  status: ResourceStatus;
  createdAt: string;
  tags: Tag[];
}
