import ExpenseTracker from "@/components/expense-tracker"

export default function Page() {
  return (
    <main className="min-h-screen bg-background py-8 px-4">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8 text-center">
          <h1 className="text-4xl font-bold text-balance mb-2">Gestión de Gastos Personales</h1>
          <p className="text-muted-foreground text-lg">Organiza y controla tus finanzas de forma simple</p>
        </div>
        <ExpenseTracker />
      </div>
    </main>
  )
}
