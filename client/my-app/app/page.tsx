'use client'

import { useState, useEffect } from 'react'
import Papa from 'papaparse'
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Search, X, Film, ThumbsUp } from 'lucide-react'

interface Movie {
  id: string
  name: string
  date: string
}

const PAGE_SIZE = 100; // Número de películas por página

export default function MovieRecommender() {
  const [movies, setMovies] = useState<Movie[]>([])
  const [selectedMovies, setSelectedMovies] = useState<Movie[]>([])
  const [recommendations, setRecommendations] = useState<Movie[]>([])
  const [searchTerm, setSearchTerm] = useState('')
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [isLoading, setIsLoading] = useState(false)
  const [allMovies, setAllMovies] = useState<Movie[]>([]);

  const loadMovies = async () => {
    setIsLoading(true);
    const response = await fetch('/movie_titles.csv');
    const csvText = await response.text();
  
    // Usamos PapaParse para convertir el CSV en un array de objetos
    Papa.parse(csvText, {
      header: false, // No usa las primeras filas como headers
      skipEmptyLines: true, // Ignorar líneas vacías
      complete: (result) => {
        const formattedMovies = (result.data as string[][]).map((row: string[]) => ({
          id: row[0],        // ID de la película
          date: row[1],      // Año de la película
          name: row[2],      // Título de la película
        }));
  
        // Guardamos todas las películas en allMovies solo una vez
        setAllMovies(formattedMovies);
  
        // Cargamos las primeras películas
        const moviesForPage = formattedMovies.slice(0, PAGE_SIZE); // Página 1
        setMovies(moviesForPage); // Mostrar las primeras películas
        setIsLoading(false);
      },
    });
  };  
  
  const connectWebSocket = () => {
    const ws = new WebSocket('ws://localhost:5902/ws');
    ws.onopen = () => console.log('Connected to WebSocket');
  
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log('Received data:', data);
  
      // Verificamos que `movieIds` sea un array
      if (Array.isArray(data.movieIds)) {
        // Mapeamos los movieIds a los nombres correspondientes usando el array `allMovies`
        const mappedRecommendations = data.movieIds.map((id: number) => {
          console.log('ID:', id);
          console.log('All Movies:', allMovies);
          console.log('All Movies Con el Id:', allMovies.map((movie) => movie.id === id.toString()));
          const movie = allMovies.find((movie) => movie.id === id.toString()); // Asegúrate de que los IDs coincidan como strings
          return movie ? movie.name : `Movie not found (${id})`;
        });

        console.log('Recommendations:', mappedRecommendations);
  
        // Actualizamos las recomendaciones con los nombres
        setRecommendations((prev) => [...prev, ...mappedRecommendations]);
      } else {
        console.error('Error: Expected an array, but got:', data);
      }
    };
  
    ws.onerror = (error) => console.log('WebSocket error:', error);
    ws.onclose = () => console.log('WebSocket connection closed');
  
    setSocket(ws);
  };

  useEffect(() => {
    loadMovies()
    connectWebSocket()
    return () => {
      if (socket) socket.close()
    }
  }, [currentPage]) // Re-cargar películas cuando cambie la página

  const handleSelectMovie = (movie: Movie) => {
    if (selectedMovies.length < 5 && !selectedMovies.some(m => m.id === movie.id)) {
      setSelectedMovies(prev => [...prev, movie])
    }
  }

  const handleRemoveMovie = (movieId: string) => {
    setSelectedMovies(prev => prev.filter(m => m.id !== movieId))
  }

  const handleSubmit = async () => {
    if (selectedMovies.length === 5) {
      try {
        const response = await fetch('http://localhost:5902/api', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ movieIds: selectedMovies.map(m => parseInt(m.id, 10)) }),
        })
        const data = await response.json()
        console.log(data)
        alert("Movies submitted successfully")
      } catch (error) {
        console.error("Error submitting movies", error)
        alert("Error submitting movies")
      }
    } else {
      alert("Please select 5 movies.")
    }
  }

  const filteredMovies = movies.filter(movie =>
    movie.name.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const loadMoreMovies = () => {
    setCurrentPage((prevPage) => prevPage + 1); // Incrementar la página actual
    const nextMovies = allMovies.slice(
      currentPage * PAGE_SIZE,              // Desplazar por la cantidad de películas mostradas
      (currentPage + 1) * PAGE_SIZE         // Mostrar la siguiente sección de películas
    );
    setMovies((prevMovies) => [...prevMovies, ...nextMovies]); // Agregar más películas a las ya mostradas
  };

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-3xl font-bold mb-6 text-center">Movie Recommender</h1>
      
      <Tabs defaultValue="select" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="select">Select Movies</TabsTrigger>
          <TabsTrigger value="recommendations">Recommendations</TabsTrigger>
        </TabsList>
        <TabsContent value="select">
          <Card>
            <CardHeader>
              <CardTitle>Select Your 5 Favorite Movies</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="mb-4">
                <div className="relative">
                  <Input
                    type="text"
                    placeholder="Search movies..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
                </div>
              </div>
              <ScrollArea className="h-[50vh]">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {filteredMovies.map((movie) => (
                    <Card key={movie.id} className="cursor-pointer hover:shadow-lg transition-shadow">
                      <CardHeader>
                        <CardTitle className="text-sm">{movie.name}</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <p className="text-xs text-gray-500">{movie.date}</p>
                      </CardContent>
                      <CardFooter>
                        <Button 
                          onClick={() => handleSelectMovie(movie)}
                          disabled={selectedMovies.length >= 5 || selectedMovies.some(m => m.id === movie.id)}
                          className="w-full text-xs"
                        >
                          {selectedMovies.some(m => m.id === movie.id) ? 'Selected' : 'Select'}
                        </Button>
                      </CardFooter>
                    </Card>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
          <div className="mt-6">
            <h2 className="text-xl font-semibold mb-4">Selected Movies ({selectedMovies.length}/5)</h2>
            <div className="flex flex-wrap gap-2 mb-4">
              {selectedMovies.map((movie) => (
                <Badge key={movie.id} variant="secondary" className="text-sm py-1 px-2">
                  {movie.name}
                  <button onClick={() => handleRemoveMovie(movie.id)} className="ml-2 text-red-500 hover:text-red-700">
                    <X size={14} />
                  </button>
                </Badge>
              ))}
            </div>
            <Button onClick={handleSubmit} disabled={selectedMovies.length !== 5} className="w-full">
              Submit Selection
            </Button>
          </div>
          <div className="mt-4">
            <Button onClick={loadMoreMovies} className="w-full" disabled={isLoading}>
              {isLoading ? 'Loading...' : 'Load More Movies'}
            </Button>
          </div>
        </TabsContent>
        <TabsContent value="recommendations">
          <Card>
            <CardHeader>
              <CardTitle>Movie Recommendations</CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-[60vh]">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {recommendations.map((movie, index) => (
                    <Card key={index}>
                      <CardHeader>
                        <CardTitle className="text-sm flex items-center">
                          <Film className="mr-2" size={16} />
                          {movie.name}
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <p className="text-xs text-gray-500">{movie.date}</p>
                      </CardContent>
                      <CardFooter>
                        <Button variant="outline" className="w-full text-xs">
                          <ThumbsUp className="mr-2" size={14} />
                          Add to Watchlist
                        </Button>
                      </CardFooter>
                    </Card>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
