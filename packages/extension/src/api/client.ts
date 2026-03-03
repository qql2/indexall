import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { TagService } from '../gen/indexall/v1/tag_connect';
import { ResourceService } from '../gen/indexall/v1/resource_connect';

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8080',
});

export const tagClient = createClient(TagService, transport);
export const resourceClient = createClient(ResourceService, transport);
