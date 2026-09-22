import { z } from 'zod';

/**
 * Health contract of the Go API. Mirrors the Go `healthResponse` DTO 1:1
 * (ADR-0006) — the schema is the only thing allowed to cross the HTTP
 * boundary.
 */
export const HealthSchema = z.object({
	status: z.enum(['ok', 'degraded']),
	database: z.enum(['ok', 'unavailable'])
});

export type Health = z.infer<typeof HealthSchema>;
