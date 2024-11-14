'use client'

import { useState, useEffect } from 'react'
import Papa from 'papaparse'
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Search, X, Loader } from 'lucide-react'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";

interface Movie {
  id: string
  name: string
  date: string
}

const PAGE_SIZE = 10 // Número de películas por página

export default function MovieRecommender() {
  const [allMovies, setAllMovies] = useState<Movie[]>([])
  const [movies, setMovies] = useState<Movie[]>([])
  const [selectedMovies, setSelectedMovies] = useState<Movie[]>([])
  const [recommendations, setRecommendations] = useState<Movie[]>([])
  const [searchTerm, setSearchTerm] = useState('')
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [isLoading, setIsLoading] = useState(false)
  const [totalPages, setTotalPages] = useState(1);

  const loadMovies = async () => {
    setIsLoading(true)
    const response = await fetch('/movie_titles.csv')
    const csvText = await response.text()

    Papa.parse(csvText, {
      header: false,
      skipEmptyLines: true,
      complete: (result) => {
        const formattedMovies = (result.data as string[][]).map((row: string[]) => ({
          id: row[0],
          date: row[1],
          name: row[2],
        }))
        setTotalPages(Math.ceil(formattedMovies.length / PAGE_SIZE));
        setAllMovies(formattedMovies)
        updateMoviesForPage(formattedMovies, currentPage)
        setIsLoading(false)
      },
    })
  }

  const updateMoviesForPage = (movieList: Movie[], page: number) => {
    const moviesForPage = movieList.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)
    setMovies(moviesForPage)
  }

  const connectWebSocket = () => {
    const ws = new WebSocket('ws://localhost:5902/ws')
    ws.onopen = () => console.log('Connected to WebSocket')
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (Array.isArray(data.movieIds)) {
        const mappedRecommendations = data.movieIds.map((id: number) => {
          const movie = allMovies.find((movie) => movie.id === id.toString())
          return movie;
        })
        console.log('Recommendations:', mappedRecommendations)
        setRecommendations(mappedRecommendations)
      } else {
        console.error('Error: Expected an array, but got:', data)
      }
    }
    ws.onerror = (error) => console.log('WebSocket error:', error)
    ws.onclose = () => console.log('WebSocket connection closed')
    setSocket(ws)
  }

  useEffect(() => {
    loadMovies()
    connectWebSocket()
    return () => {
      if (socket) socket.close()
    }
  }, [])

  useEffect(() => {
    updateMoviesForPage(allMovies, currentPage)
  }, [currentPage, allMovies])

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

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    setIsLoading(true);
    setTimeout(() => setIsLoading(false), 500); // Simulación de carga
  }

  // Configuración de las páginas que se muestran
  const visibleRange = 1; // Número de páginas que deseas mostrar alrededor de la página actual
  const pages = [];

  for (let i = 1; i <= totalPages; i++) {
      if (
          i === 1 || // Siempre muestra la primera página
          i === totalPages || // Siempre muestra la última página
          (i >= currentPage - visibleRange && i <= currentPage + visibleRange) // Muestra un rango alrededor de la página actual
      ) {
          pages.push(i);
      } else if (
          (i === currentPage - visibleRange - 1 || i === currentPage + visibleRange + 1) &&
          pages[pages.length - 1] !== '...'
      ) {
          pages.push('...'); // Agrega un elipsis si hay páginas ocultas
      }
  }

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-3xl font-bold mb-6 text-center">Sistemas de Recomendación de Películas</h1>

      <Tabs defaultValue="select" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="select">Seleccionar Películas</TabsTrigger>
          <TabsTrigger value="recommendations">Recomendaciones</TabsTrigger>
        </TabsList>
        <TabsContent value="select">
          <Card>
            <CardHeader>
              <CardTitle>Selecciona tus 5 películas favoritas</CardTitle>
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
                {isLoading ? (
                  <div className="flex justify-center items-center h-full">
                    <Loader className="animate-spin text-gray-500" size={48} />
                  </div>
                ) : (
                  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-4">
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
                )}
              </ScrollArea>
            </CardContent>
          </Card>
          {/* Paginación */}
          <Pagination className="mt-4 flex justify-center items-center space-x-2">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  href="#"
                  onClick={(e) => {
                    e.preventDefault();
                    if (currentPage > 1) handlePageChange(currentPage - 1);
                  }}
                  className={currentPage === 1 ? 'disabled' : ''}
                >
                  Anterior
                </PaginationPrevious>
              </PaginationItem>

              {pages.map((page, index) =>
                page === '...' ? (
                  <PaginationEllipsis key={index} />
                ) : (
                  <PaginationItem key={index} className={page === currentPage ? 'active' : ''}>
                    <PaginationLink
                      href="#"
                      onClick={(e) => {
                        e.preventDefault();
                        handlePageChange(page as number);
                      }}
                    >
                      {page}
                    </PaginationLink>
                  </PaginationItem>
                )
              )}

              <PaginationItem>
                <PaginationNext
                  href="#"
                  onClick={(e) => {
                    e.preventDefault();
                    if (currentPage < totalPages) handlePageChange(currentPage + 1);
                  }}
                  className={currentPage === totalPages ? 'disabled' : ''}
                >
                  Siguiente
                </PaginationNext>
              </PaginationItem>
            </PaginationContent>
          </Pagination>
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
            <Button onClick={handleSubmit} disabled={selectedMovies.length !== 5}>
              Submit Movies
            </Button>
          </div>
        </TabsContent>

        <TabsContent value="recommendations">
          <h2 className="text-xl font-semibold mb-4">Recommended Movies</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {recommendations.map((movie) => (
              <Card key={movie.id}>
                <CardHeader>
                  <CardTitle className="text-sm">{movie.name}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-xs text-gray-500">{movie.date}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}