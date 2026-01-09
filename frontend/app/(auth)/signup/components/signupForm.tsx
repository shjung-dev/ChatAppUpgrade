"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function SignUpForm() {
  const router = useRouter()
  const API_BASE = "https://upgradedchatappservice.onrender.com";

  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)

    const payload = {
      username,
      password,
    }

    try {
      const res = await fetch(`${API_BASE}/signup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      })

      const data = await res.json()

      if (!res.ok) {
        throw new Error(data.error || "Signup failed")
      }

      if (
        data.access_token &&
        data.refresh_token &&
        data.message === "signup successful"
      ) {
        sessionStorage.setItem("access_token", data.access_token)
        sessionStorage.setItem("refresh_token", data.refresh_token)
        sessionStorage.setItem("username", data.user.username)
      }

      router.push("/home")

    } catch (err: any) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card className="mx-auto max-w-sm rounded-2xl border border-border/60 bg-background/80 backdrop-blur shadow-xl transition-all duration-300 hover:shadow-2xl">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl font-semibold tracking-tight">
          Create your account
        </CardTitle>
        <CardDescription className="text-muted-foreground">
          Sign up with a username and password
        </CardDescription>
      </CardHeader>

      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-4">

            <div className="grid gap-2">
              <Label htmlFor="username" className="text-sm font-medium">
                Username
              </Label>
              <Input
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                className="h-11 rounded-xl transition-all focus-visible:ring-2 focus-visible:ring-primary"
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="password" className="text-sm font-medium">
                Password
              </Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="h-11 rounded-xl transition-all focus-visible:ring-2 focus-visible:ring-primary"
              />
            </div>

            {error && (
              <p className="text-sm text-red-500 animate-in fade-in slide-in-from-top-1">
                {error}
              </p>
            )}

            <Button
              type="submit"
              className="h-11 w-full rounded-xl text-base font-medium transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]"
              disabled={loading}
            >
              {loading ? "Creating account..." : "Create an account"}
            </Button>
          </div>
        </form>

        <div className="mt-6 text-center text-sm text-muted-foreground">
          Already have an account?{" "}
          <Link
            href="/"
            className="font-medium text-primary underline-offset-4 transition-colors hover:underline"
          >
            Sign in
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
