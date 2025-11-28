"use client"

import { useEffect, useMemo, useState } from "react"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Spinner } from "@/components/ui/spinner"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Calendar, DollarSign, LogOut, Plus, Repeat, Tag, Trash2 } from "lucide-react"
import {
  endOfMonth,
  endOfWeek,
  endOfYear,
  format,
  isWithinInterval,
  parseISO,
  startOfMonth,
  startOfWeek,
  startOfYear,
} from "date-fns"
import { es } from "date-fns/locale"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api"

interface Expense {
  id: number
  name: string
  tag: string
  amount: number
  date: string
}

interface RecurringExpense {
  id: number
  name: string
  tag: string
  amount: number
  lastAppliedAt?: string | null
  lastExpenseId?: number | null
}

interface AuthUser {
  id: number
  name: string
  email: string
}

type Period = "day" | "week" | "month" | "year" | "custom"

async function fetchJSON<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const bodyProvided = options.body !== undefined
  const headers = new Headers(options.headers)
  const method = (options.method ?? "GET").toUpperCase()

  if (bodyProvided && !(options.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers })
  } catch (error) {
    const reason = error instanceof Error ? error.message : "Error desconocido"
    throw new Error(`No se pudo contactar al servidor (${method} ${path}): ${reason}`)
  }

  if (!response.ok) {
    let message = ""
    let detail = ""
    const isJSON = response.headers.get("content-type")?.includes("application/json")
    if (isJSON) {
      try {
        const data = await response.json()
        if (typeof data?.error === "string") {
          message = data.error
        }
        if (typeof data?.details === "string" && data.details.trim().length > 0) {
          detail = data.details
        }
      } catch {
        // ignore parsing errors, fallback later
      }
    }

    if (!message) {
      message = `El servidor respondió ${response.status} al intentar ${method} ${path}`
    }
    if (detail) {
      message = `${message} (${detail})`
    }

    throw new Error(message)
  }

  if (response.status === 204) {
    return null as T
  }

  const contentType = response.headers.get("content-type") ?? ""
  if (!contentType.includes("application/json")) {
    return null as T
  }

  return (await response.json()) as T
}

const canApplyRecurringExpense = (recurring: RecurringExpense) => {
  if (!recurring.lastAppliedAt) return true
  try {
    const lastDate = parseISO(recurring.lastAppliedAt)
    const now = new Date()
    return !(lastDate.getFullYear() === now.getFullYear() && lastDate.getMonth() === now.getMonth())
  } catch {
    return true
  }
}

export default function ExpenseTracker() {
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [recurringExpenses, setRecurringExpenses] = useState<RecurringExpense[]>([])
  const [name, setName] = useState("")
  const [tag, setTag] = useState("")
  const [amount, setAmount] = useState("")
  const [date, setDate] = useState(format(new Date(), "yyyy-MM-dd"))
  const [filterPeriod, setFilterPeriod] = useState<Period>("month")
  const [customStartDate, setCustomStartDate] = useState("")
  const [customEndDate, setCustomEndDate] = useState("")
  const [recurringName, setRecurringName] = useState("")
  const [recurringTag, setRecurringTag] = useState("")
  const [recurringAmount, setRecurringAmount] = useState("")

  const [token, setToken] = useState<string | null>(null)
  const [user, setUser] = useState<AuthUser | null>(null)
  const [initializing, setInitializing] = useState(true)
  const [authMode, setAuthMode] = useState<"login" | "register">("login")
  const [authForm, setAuthForm] = useState({ name: "", email: "", password: "" })
  const [authLoading, setAuthLoading] = useState(false)
  const [authError, setAuthError] = useState<string | null>(null)
  const [globalError, setGlobalError] = useState<string | null>(null)
  const [syncing, setSyncing] = useState(false)

  useEffect(() => {
    if (typeof window === "undefined") return
    const storedToken = localStorage.getItem("authToken")
    if (!storedToken) {
      setInitializing(false)
      return
    }
    initializeSession(storedToken)
  }, [])

  const initializeSession = async (authToken: string, bootstrapUser?: AuthUser) => {
    setGlobalError(null)
    try {
      localStorage.setItem("authToken", authToken)
      setToken(authToken)

      if (bootstrapUser) {
        setUser(bootstrapUser)
      } else {
        const me = await fetchJSON<{ user: AuthUser }>("/auth/me", {}, authToken)
        setUser(me.user)
      }

      await fetchAllData(authToken)
    } catch (error) {
      console.error(error)
      setGlobalError(error instanceof Error ? error.message : "No se pudo iniciar sesión")
      localStorage.removeItem("authToken")
      setToken(null)
      setUser(null)
    } finally {
      setInitializing(false)
    }
  }

  const fetchAllData = async (authToken: string) => {
    setSyncing(true)
    try {
      const [expensesResponse, recurringResponse] = await Promise.all([
        fetchJSON<{ expenses: Expense[] }>("/expenses", {}, authToken),
        fetchJSON<{ monthlyExpenses: RecurringExpense[] }>("/monthly-expenses", {}, authToken),
      ])
      setExpenses(expensesResponse?.expenses ?? [])
      setRecurringExpenses(recurringResponse?.monthlyExpenses ?? [])
    } catch (error) {
      console.error(error)
      setGlobalError(error instanceof Error ? error.message : "No se pudo sincronizar con el servidor")
    } finally {
      setSyncing(false)
    }
  }

  const handleAuthSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setAuthLoading(true)
    setAuthError(null)

    try {
      const endpoint = authMode === "login" ? "/auth/login" : "/auth/register"
      const payload =
        authMode === "login"
          ? { email: authForm.email, password: authForm.password }
          : { name: authForm.name, email: authForm.email, password: authForm.password }

      const response = await fetchJSON<{ token: string; user: AuthUser }>(endpoint, {
        method: "POST",
        body: JSON.stringify(payload),
      })

      await initializeSession(response.token, response.user)
      setAuthForm({ name: "", email: "", password: "" })
    } catch (error) {
      setAuthError(error instanceof Error ? error.message : "No se pudo completar la acción")
    } finally {
      setAuthLoading(false)
    }
  }

  const handleLogout = () => {
    localStorage.removeItem("authToken")
    setToken(null)
    setUser(null)
    setExpenses([])
    setRecurringExpenses([])
  }

  const addExpense = async () => {
    if (!token || !name || !tag || !amount || !date) return
    const parsedAmount = Number.parseFloat(amount)
    if (Number.isNaN(parsedAmount)) return

    try {
      const response = await fetchJSON<{ expense: Expense }>("/expenses", {
        method: "POST",
        body: JSON.stringify({ name, tag, amount: parsedAmount, date }),
      }, token)

      if (response?.expense) {
        setExpenses((prev) => [...prev, response.expense])
      }

      setName("")
      setTag("")
      setAmount("")
      setDate(format(new Date(), "yyyy-MM-dd"))
    } catch (error) {
      setGlobalError(error instanceof Error ? error.message : "No se pudo crear el gasto")
    }
  }

  const addRecurringExpense = async () => {
    if (!token || !recurringName || !recurringTag || !recurringAmount) return
    const parsedAmount = Number.parseFloat(recurringAmount)
    if (Number.isNaN(parsedAmount)) return

    try {
      const response = await fetchJSON<{ monthlyExpense: RecurringExpense }>("/monthly-expenses", {
        method: "POST",
        body: JSON.stringify({ name: recurringName, tag: recurringTag, amount: parsedAmount }),
      }, token)

      if (response?.monthlyExpense) {
        setRecurringExpenses((prev) => [response.monthlyExpense, ...prev])
      }

      setRecurringName("")
      setRecurringTag("")
      setRecurringAmount("")
    } catch (error) {
      setGlobalError(error instanceof Error ? error.message : "No se pudo crear el gasto recurrente")
    }
  }

  const applyRecurringExpense = async (recurring: RecurringExpense) => {
    if (!token) return
    if (!canApplyRecurringExpense(recurring)) {
      setGlobalError("Este gasto recurrente ya fue aplicado este mes")
      return
    }
    try {
      const response = await fetchJSON<{ expense: Expense; monthlyExpense: RecurringExpense }>(
        `/monthly-expenses/${recurring.id}/apply`,
        {
          method: "POST",
        },
        token,
      )

      if (response?.expense) {
        setExpenses((prev) => [...prev, response.expense])
      }
      if (response?.monthlyExpense) {
        setRecurringExpenses((prev) => prev.map((item) => (item.id === recurring.id ? response.monthlyExpense : item)))
      }
    } catch (error) {
      setGlobalError(error instanceof Error ? error.message : "No se pudo aplicar el gasto recurrente")
    }
  }

  const deleteExpense = async (id: number) => {
    if (!token) return
    try {
      await fetchJSON<null>(`/expenses/${id}`, { method: "DELETE" }, token)
      setExpenses((prev) => prev.filter((expense) => expense.id !== id))
      setRecurringExpenses((prev) =>
        prev.map((recurring) =>
          recurring.lastExpenseId === id
            ? { ...recurring, lastExpenseId: null, lastAppliedAt: null }
            : recurring,
        ),
      )
    } catch (error) {
      setGlobalError(error instanceof Error ? error.message : "No se pudo eliminar el gasto")
    }
  }

  const deleteRecurringExpense = async (id: number) => {
    if (!token) return
    try {
      await fetchJSON<null>(`/monthly-expenses/${id}`, { method: "DELETE" }, token)
      setRecurringExpenses((prev) => prev.filter((expense) => expense.id !== id))
    } catch (error) {
      setGlobalError(error instanceof Error ? error.message : "No se pudo eliminar el gasto recurrente")
    }
  }

  const availableTags = useMemo(() => {
    const tags = new Set<string>()
    expenses.forEach((expense) => tags.add(expense.tag))
    recurringExpenses.forEach((expense) => tags.add(expense.tag))
    return Array.from(tags)
  }, [expenses, recurringExpenses])

  const getFilteredExpenses = () => {
    const now = new Date()

    return expenses.filter((expense) => {
      const expenseDate = parseISO(expense.date)
      let rangeStart: Date | null = null
      let rangeEnd: Date | null = null

      switch (filterPeriod) {
        case "day":
          return format(expenseDate, "yyyy-MM-dd") === format(now, "yyyy-MM-dd")
        case "week":
          rangeStart = startOfWeek(now, { weekStartsOn: 1 })
          rangeEnd = endOfWeek(now, { weekStartsOn: 1 })
          break
        case "month":
          rangeStart = startOfMonth(now)
          rangeEnd = endOfMonth(now)
          break
        case "year":
          rangeStart = startOfYear(now)
          rangeEnd = endOfYear(now)
          break
        case "custom":
          if (!customStartDate || !customEndDate) return true
          rangeStart = parseISO(customStartDate)
          rangeEnd = parseISO(customEndDate)
          break
        default:
          return true
      }

      if (!rangeStart || !rangeEnd) return true

      return isWithinInterval(expenseDate, {
        start: rangeStart,
        end: rangeEnd,
      })
    })
  }

  const filteredExpenses = getFilteredExpenses()
  const total = filteredExpenses.reduce((sum, expense) => sum + expense.amount, 0)

  const groupedByTag = filteredExpenses.reduce(
    (acc, expense) => {
      if (!acc[expense.tag]) {
        acc[expense.tag] = 0
      }
      acc[expense.tag] += expense.amount
      return acc
    },
    {} as Record<string, number>,
  )

  if (initializing) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Spinner className="h-5 w-5" />
          <span>Cargando...</span>
        </div>
      </div>
    )
  }

  if (!token || !user) {
    return (
      <Card className="max-w-md mx-auto">
        <CardHeader>
          <CardTitle>{authMode === "login" ? "Inicia Sesión" : "Crea tu cuenta"}</CardTitle>
        </CardHeader>
        <CardContent>
          {authError && (
            <Alert className="mb-4" variant="destructive">
              <AlertDescription>{authError}</AlertDescription>
            </Alert>
          )}
          <form className="space-y-4" onSubmit={handleAuthSubmit}>
            {authMode === "register" && (
              <div className="space-y-2">
                <Label htmlFor="name">Nombre</Label>
                <Input
                  id="name"
                  value={authForm.name}
                  onChange={(event) => setAuthForm((prev) => ({ ...prev, name: event.target.value }))}
                  required
                />
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={authForm.email}
                onChange={(event) => setAuthForm((prev) => ({ ...prev, email: event.target.value }))}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Contraseña</Label>
              <Input
                id="password"
                type="password"
                value={authForm.password}
                onChange={(event) => setAuthForm((prev) => ({ ...prev, password: event.target.value }))}
                minLength={6}
                required
              />
            </div>

            <Button type="submit" className="w-full" disabled={authLoading}>
              {authLoading ? (
                <span className="flex items-center gap-2">
                  <Spinner className="h-4 w-4" />
                  Procesando...
                </span>
              ) : authMode === "login" ? (
                "Ingresar"
              ) : (
                "Registrarme"
              )}
            </Button>
          </form>
          <p className="mt-4 text-center text-sm text-muted-foreground">
            {authMode === "login" ? "¿Todavía no tienes cuenta?" : "¿Ya tienes una cuenta?"}{" "}
            <button
              type="button"
              className="font-medium text-primary hover:underline"
              onClick={() => {
                setAuthMode(authMode === "login" ? "register" : "login")
                setAuthError(null)
              }}
            >
              {authMode === "login" ? "Regístrate" : "Inicia sesión"}
            </button>
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-sm text-muted-foreground">Hola, {user.name}</p>
          <p className="text-lg font-semibold">{user.email}</p>
        </div>
        <Button variant="outline" size="sm" onClick={handleLogout} className="w-full lg:w-auto">
          <LogOut className="mr-2 h-4 w-4" />
          Cerrar sesión
        </Button>
      </div>

      {globalError && (
        <Alert variant="destructive">
          <AlertDescription>{globalError}</AlertDescription>
        </Alert>
      )}

      {syncing && (
        <div className="flex items-center gap-2 rounded-md border border-dashed border-muted px-3 py-2 text-sm text-muted-foreground">
          <Spinner className="h-4 w-4" />
          Sincronizando con el backend...
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Resumen de Gastos</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs value={filterPeriod} onValueChange={(v) => setFilterPeriod(v as any)}>
            <TabsList className="grid w-full grid-cols-5 mb-6">
              <TabsTrigger value="day">Día</TabsTrigger>
              <TabsTrigger value="week">Semana</TabsTrigger>
              <TabsTrigger value="month">Mes</TabsTrigger>
              <TabsTrigger value="year">Año</TabsTrigger>
              <TabsTrigger value="custom">Custom</TabsTrigger>
            </TabsList>

            <TabsContent value="custom" className="mt-0">
              <div className="grid grid-cols-2 gap-4 mb-6">
                <div className="space-y-2">
                  <Label htmlFor="start-date">Fecha inicio</Label>
                  <Input
                    id="start-date"
                    type="date"
                    value={customStartDate}
                    onChange={(e) => setCustomStartDate(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="end-date">Fecha fin</Label>
                  <Input
                    id="end-date"
                    type="date"
                    value={customEndDate}
                    onChange={(e) => setCustomEndDate(e.target.value)}
                  />
                </div>
              </div>
            </TabsContent>
          </Tabs>

          <div className="mb-6 p-6 bg-accent/20 rounded-xl border-2 border-accent">
            <p className="text-sm text-muted-foreground mb-1">Total del período</p>
            <p className="text-4xl font-bold text-accent-foreground">${total.toFixed(2)}</p>
          </div>

          {Object.keys(groupedByTag).length > 0 && (
            <div className="mb-6">
              <h3 className="text-sm font-semibold mb-3">Por Etiqueta</h3>
              <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
                {Object.entries(groupedByTag).map(([tag, amount]) => (
                  <div key={tag} className="flex items-center justify-between p-3 bg-secondary rounded-lg">
                    <Badge variant="outline" className="font-medium">
                      <Tag className="h-3 w-3 mr-1" />
                      {tag}
                    </Badge>
                    <span className="font-semibold">${amount.toFixed(2)}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="space-y-3">
            <h3 className="text-sm font-semibold">Gastos ({filteredExpenses.length})</h3>
            {filteredExpenses.length === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                <p>No hay gastos en este período</p>
              </div>
            ) : (
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {filteredExpenses
                  .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
                  .map((expense) => (
                    <div
                      key={expense.id}
                      className="flex items-center justify-between p-4 bg-card border rounded-lg hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <p className="font-semibold">{expense.name}</p>
                          <Badge variant="secondary" className="text-xs">
                            {expense.tag}
                          </Badge>
                        </div>
                        <p className="text-xs text-muted-foreground">
                          {format(parseISO(expense.date), "d 'de' MMMM, yyyy", { locale: es })}
                        </p>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="text-lg font-bold">${expense.amount.toFixed(2)}</span>
                        <Button variant="ghost" size="icon" onClick={() => deleteExpense(expense.id)} className="text-destructive hover:text-destructive">
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Plus className="h-5 w-5" />
              Agregar Gasto
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Nombre</Label>
              <Input id="name" placeholder="Ej: Supermercado" value={name} onChange={(e) => setName(e.target.value)} />
            </div>

            <div className="space-y-2">
              <Label htmlFor="tag">Etiqueta</Label>
              <Input
                id="tag"
                placeholder="Escribe o selecciona"
                value={tag}
                onChange={(e) => setTag(e.target.value)}
                list="all-tags"
              />
              <datalist id="all-tags">
                {availableTags.map((t) => (
                  <option key={t} value={t} />
                ))}
              </datalist>
            </div>

            <div className="space-y-2">
              <Label htmlFor="amount">Monto</Label>
              <div className="relative">
                <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="amount"
                  type="number"
                  step="0.01"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="pl-9"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="date">Fecha</Label>
              <div className="relative">
                <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input id="date" type="date" value={date} onChange={(e) => setDate(e.target.value)} className="pl-9" />
              </div>
            </div>

            <Button onClick={addExpense} className="w-full">
              <Plus className="mr-2 h-4 w-4" />
              Agregar
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Repeat className="h-5 w-5" />
              Gastos Recurrentes Mensuales
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="recurring-name">Nombre</Label>
                <Input
                  id="recurring-name"
                  placeholder="Ej: Luz, Agua, Internet"
                  value={recurringName}
                  onChange={(e) => setRecurringName(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="recurring-tag">Etiqueta</Label>
                <Input
                  id="recurring-tag"
                  placeholder="Escribe o selecciona"
                  value={recurringTag}
                  onChange={(e) => setRecurringTag(e.target.value)}
                  list="tags-list"
                />
                <datalist id="tags-list">
                  {availableTags.map((t) => (
                    <option key={t} value={t} />
                  ))}
                </datalist>
              </div>

              <div className="space-y-2">
                <Label htmlFor="recurring-amount">Monto</Label>
                <div className="relative">
                  <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="recurring-amount"
                    type="number"
                    step="0.01"
                    placeholder="0.00"
                    value={recurringAmount}
                    onChange={(e) => setRecurringAmount(e.target.value)}
                    className="pl-9"
                  />
                </div>
              </div>

              <Button onClick={addRecurringExpense} className="w-full">
                <Plus className="mr-2 h-4 w-4" />
                Agregar Recurrente
              </Button>
            </div>

            {recurringExpenses.length > 0 && (
              <div className="space-y-2 pt-4 border-t">
                <h3 className="text-sm font-semibold mb-2">Mis Gastos Recurrentes</h3>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {recurringExpenses.map((recurring) => (
                    <div
                      key={recurring.id}
                      className="flex items-center justify-between p-3 bg-muted/50 rounded-lg text-sm"
                    >
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-medium">{recurring.name}</span>
                          <Badge variant="outline" className="text-xs">
                            {recurring.tag}
                          </Badge>
                        </div>
                        <span className="text-sm font-semibold">${recurring.amount.toFixed(2)}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => applyRecurringExpense(recurring)}
                          className="h-8"
                          disabled={!canApplyRecurringExpense(recurring)}
                        >
                          {canApplyRecurringExpense(recurring) ? "Aplicar" : "Aplicado"}
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8"
                          onClick={() => deleteRecurringExpense(recurring.id)}
                        >
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
