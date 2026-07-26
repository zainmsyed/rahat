import '@testing-library/jest-dom';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import InfoBox from './InfoBox.svelte';
import InfoBoxWrapper from './InfoBoxWrapper.svelte';

describe('InfoBox', () => {
	it('renders a title', () => {
		render(InfoBox, { props: { title: 'Heads up' } });
		expect(screen.getByText('Heads up')).toBeInTheDocument();
	});

	it('renders slot content through a wrapper', () => {
		render(InfoBoxWrapper, { props: { title: 'Heads up' } });
		expect(screen.getByText('Heads up')).toBeInTheDocument();
		expect(screen.getByText('This is informational.')).toBeInTheDocument();
	});
});
