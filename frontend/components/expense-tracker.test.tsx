import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import ExpenseTracker from './expense-tracker'

describe('ExpenseTracker', () => {
    it('renders login screen by default', () => {
        render(<ExpenseTracker />)
        // Since there is no token in localStorage, it should show the login card
        expect(screen.getByText('Inicia Sesión')).toBeDefined()
    })
})
