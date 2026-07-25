// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		interface SessionUser {
			id: string;
			display_name: string;
			timezone: string;
			daily_time_budget_minutes: number;
			email: string;
		}
		interface Locals {
			sessionUser: SessionUser | null;
		}
	}
}

export {};
