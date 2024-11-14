'use client'

import { useState, useEffect } from 'react'
import { Moon, Sun, Send } from 'lucide-react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

type Message = {
  content: string
}

export default function Component() {
  const [messages, setMessages] = useState<Message[]>([])
  const [content, setContent] = useState<string>('')
  const [isDarkMode, setIsDarkMode] = useState<boolean>(false)

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:5902/ws')

    socket.onopen = () => {
      console.log('Connected to WebSocket server')
    }

    socket.onmessage = (event) => {
      const newMessage: Message = JSON.parse(event.data)
      setMessages((prevMessages) => [...prevMessages, newMessage])
    }

    socket.onerror = (error) => {
      console.log('WebSocket connection error', error)
    }

    return () => {
      socket.close()
    }
  }, [])

  const sendMessage = () => {
    if (content.trim()) {
      const socket = new WebSocket('ws://localhost:5902/api')
      socket.onopen = () => {
        socket.send(JSON.stringify({ content }))
        setContent('')
      }
    }
  }

  const toggleTheme = () => {
    setIsDarkMode(!isDarkMode)
  }

  return (
    <div className={`min-h-screen ${isDarkMode ? 'dark' : ''}`}>
      <div className="container mx-auto p-4 bg-background text-foreground">
        <Card className="w-full max-w-2xl mx-auto">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>WebSocket Chat</CardTitle>
            <Button variant="ghost" size="icon" onClick={toggleTheme}>
              {isDarkMode ? <Sun className="h-[1.2rem] w-[1.2rem]" /> : <Moon className="h-[1.2rem] w-[1.2rem]" />}
            </Button>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="h-[400px] overflow-y-auto border rounded-md p-4">
                {messages.map((message, index) => (
                  <div key={index} className="mb-2 p-2 bg-secondary rounded-md">
                    {message.content}
                  </div>
                ))}
              </div>
              <div className="flex space-x-2">
                <Input
                  type="text"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="Type your message"
                  onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
                />
                <Button onClick={sendMessage}>
                  <Send className="h-4 w-4 mr-2" />
                  Send
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}