// Public surface of the health domain. `api/`, `queries/` and `schemas/` stay
// internal — other domains and routes import from here only.
//
// `Beranda` is the screen: it arranges the two modules and owns the title strip,
// so the route stays an adapter that only puts a component on a URL (ADR-0006).
// `HealthStatus` stays exported as the status module on its own.
export { default as Beranda } from './components/Beranda.svelte';
export { default as HealthStatus } from './components/HealthStatus.svelte';
export { createHealthQuery, healthKeys } from './queries/health.queries';
export { HealthSchema, type Health } from './schemas/health.schema';
