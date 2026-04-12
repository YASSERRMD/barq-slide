import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";

// The API base URL is injected at build time via NEXT_PUBLIC_API_BASE_URL.
// Falls back to localhost for local development.
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

/**
 * A reusable Connect transport configured for the barq-slides backend.
 * Uses fetch-based transport compatible with Next.js App Router (server + client components).
 */
export const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
  // Use the Connect protocol (JSON over HTTP/1.1 + HTTP/2).
  // Switch to grpcWebTransport if a gRPC-Web proxy is placed in front.
});

/**
 * Type-safe Connect client factory. Pass in a generated service descriptor.
 *
 * @example
 * import { HeliosService } from "@/gen/barq/v1/service_connect";
 * const helios = createRPCClient(HeliosService);
 */
export function createRPCClient<T extends object>(
  service: T
): ReturnType<typeof createClient<T>> {
  return createClient(service, transport);
}
