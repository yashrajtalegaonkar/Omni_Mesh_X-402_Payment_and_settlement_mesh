import React, { useState } from 'react';
import { Play } from 'lucide-react';

export const MOVIES = [
  {
    id: 'mov-1',
    title: 'The Algorand Enigma',
    genre: 'Thriller · Sci-Fi',
    year: 2024,
    rating: 'PG-13',
    duration: '2h 14m',
    description: 'A rogue AI embedded in the Algorand blockchain begins rewriting transaction history. One analyst discovers a pattern — and becomes the chain\'s last line of defense.',
    image: 'https://images.unsplash.com/photo-1626814026160-2237a95fc5a0?w=1200&auto=format&fit=crop&q=80',
    thumb: 'https://images.unsplash.com/photo-1626814026160-2237a95fc5a0?w=500&auto=format&fit=crop&q=60',
  },
  {
    id: 'mov-2',
    title: 'Cyberpunk 2077: Edgerunners',
    genre: 'Action · Cyberpunk',
    year: 2023,
    rating: 'R',
    duration: '1h 58m',
    description: 'Night City. A city of shattered dreams and neon promises. A street kid fights to survive in a world of corpo assassins and chrome-laced violence.',
    image: 'https://images.unsplash.com/photo-1605810230434-7631ac76ec81?w=1200&auto=format&fit=crop&q=80',
    thumb: 'https://images.unsplash.com/photo-1605810230434-7631ac76ec81?w=500&auto=format&fit=crop&q=60',
  },
  {
    id: 'mov-3',
    title: 'Deep Space Nine',
    genre: 'Drama · Space Opera',
    year: 2025,
    rating: 'PG',
    duration: '2h 31m',
    description: 'On the frontier of explored space, a crew of unlikely heroes must keep the peace between the Federation, the Bajoran people, and an ancient wormhole intelligence.',
    image: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=1200&auto=format&fit=crop&q=80',
    thumb: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=500&auto=format&fit=crop&q=60',
  },
  {
    id: 'mov-4',
    title: 'The Last Ledger',
    genre: 'Mystery · Finance',
    year: 2024,
    rating: 'PG-13',
    duration: '1h 47m',
    description: 'A forensic accountant unravels a decade-long conspiracy hidden in a blockchain ledger — where every transaction is a clue, and every block conceals a body.',
    image: 'https://images.unsplash.com/photo-1639762681485-074b7f938ba0?w=1200&auto=format&fit=crop&q=80',
    thumb: 'https://images.unsplash.com/photo-1639762681485-074b7f938ba0?w=500&auto=format&fit=crop&q=60',
  },
];

export default function MovieGrid({ selectedMovieId, onSelect }) {
  return (
    <div className="px-12 py-8">
      <h2 className="text-2xl font-bold mb-6 tracking-wide">Trending Now</h2>
      <div className="grid grid-cols-2 sm:grid-cols-2 md:grid-cols-4 gap-4">
        {MOVIES.map((movie) => {
          const isSelected = movie.id === selectedMovieId;
          return (
            <div
              key={movie.id}
              onClick={() => onSelect(movie)}
              className={`relative group rounded-lg overflow-hidden cursor-pointer transition-all duration-300 shadow-lg
                ${isSelected ? 'ring-2 ring-[#E50914] scale-[1.03] z-10' : 'hover:scale-[1.03] hover:z-10'}`}
            >
              <img src={movie.thumb} alt={movie.title} className="w-full h-52 object-cover" />
              <div className="absolute inset-0 bg-gradient-to-t from-black/90 via-black/20 to-transparent" />
              <div className="absolute bottom-0 left-0 right-0 p-3">
                <p className="text-sm font-bold text-white leading-tight">{movie.title}</p>
                <p className="text-xs text-gray-400 mt-1">{movie.genre}</p>
              </div>
              {isSelected && (
                <div className="absolute top-3 right-3 bg-[#E50914] rounded-full px-2 py-0.5 text-xs font-bold">
                  Selected
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
