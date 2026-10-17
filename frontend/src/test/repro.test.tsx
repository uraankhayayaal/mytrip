import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// Минимальный репро: проверяем, что fireEvent.change меняет значение input.
function Form() {
  const [v, setV] = React.useState('');
  return (
    <form>
      <input aria-label="field" value={v} onChange={(e) => setV(e.target.value)} />
      <button type="button" onClick={() => console.log('VALUE=' + v)}>
        submit
      </button>
      <div data-testid="v">{v}</div>
    </form>
  );
}

import React from 'react';

describe('repro', () => {
  it('change sets value', () => {
    render(<Form />);
    fireEvent.change(screen.getByLabelText('field'), { target: { value: 'abc' } });
    expect((screen.getByLabelText('field') as HTMLInputElement).value).toBe('abc');
  });
});
