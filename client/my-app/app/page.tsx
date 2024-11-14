'use client';  // Asegúrate de que este componente se renderice en el cliente

import { useState, useEffect } from 'react';

type Message = {
  content: string;
};

export default function Home() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [content, setContent] = useState<string>('');

  useEffect(() => {
    // Crea la conexión WebSocket cuando se carga el componente
    const socket = new WebSocket('ws://localhost:8080/ws');

    // Cuando se abre la conexión
    socket.onopen = () => {
      console.log('Conectado al servidor WebSocket');
    };

    // Cuando se recibe un mensaje del servidor
    socket.onmessage = (event) => {
      const newMessage: Message = JSON.parse(event.data);
      setMessages((prevMessages) => [...prevMessages, newMessage]);
    };

    // Maneja errores
    socket.onerror = (error) => {
      console.log('Error en la conexión WebSocket', error);
    };

    // Limpia la conexión WebSocket cuando se desmonte el componente
    return () => {
      socket.close();
    };
  }, []);

  // Enviar mensaje al servidor WebSocket
  const sendMessage = () => {
    const socket = new WebSocket('ws://localhost:8080/ws');
    socket.onopen = () => {
      socket.send(JSON.stringify({ content }));
      setContent(''); // Limpiar el campo de entrada después de enviar el mensaje
    };
  };

  return (
    <div>
      <h1>Conexión WebSocket en Next.js</h1>

      <div>
        <input
          type="text"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Escribe tu mensaje"
        />
        <button onClick={sendMessage}>Enviar</button>
      </div>

      <h2>Mensajes recibidos:</h2>
      <ul>
        {messages.map((message, index) => (
          <li key={index}>{message.content}</li>
        ))}
      </ul>
    </div>
  );
}
