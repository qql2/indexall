import { createConnectTransport } from '@connectrpc/connect-web';
import { createPromiseClient } from '@connectrpc/connect';
import { TagService } from '../gen/indexall/v1/tag_pb';
import { ResourceService } from '../gen/indexall/v1/resource_pb';

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8080',
});

export const tagClient = createPromiseClient(TagService, transport);
export const resourceClient = createPromiseClient(ResourceService, transport);
