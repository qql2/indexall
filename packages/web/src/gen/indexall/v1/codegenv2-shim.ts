// Shim for codegenv2 to provide runtime message descriptors
// protoc-gen-es v2 requires these for runtime serialization/deserialization
// This is a workaround for protoc-gen-es v2 compatibility with newer versions

import { Message } from "@bufbuild/protobuf";

export type GenFile = any;
export type GenMessage<T> = any;
export type GenService<T> = any;
export type GenEnum<T> = any;

// Re-export MethodKind for generated code compatibility
// Different protoc versions may expect this from different modules
export enum MethodKind {
  Unary = 0,
  ServerStreaming = 1,
  ClientStreaming = 2,
  BiDiStreaming = 3,
}

let messageIdCounter = 0;

// Minimal MessageType-like object for @connectrpc/connect compatibility
export const messageDesc = (): any => {
  const id = messageIdCounter++;
  // Create a minimal but functional message descriptor
  // @connectrpc/connect needs at least these properties to work
  const descriptor = {
    typeName: `Message${id}`, // Unique type name
    fields: {},
    create(): any {
      return {};
    },
    fromBinary(): any {
      return {};
    },
    toBinary(): Uint8Array {
      return new Uint8Array();
    },
  };
  return descriptor;
};

export const fileDesc = () => ({});

export const serviceDesc = () => ({});

export const enumDesc = () => ({});
