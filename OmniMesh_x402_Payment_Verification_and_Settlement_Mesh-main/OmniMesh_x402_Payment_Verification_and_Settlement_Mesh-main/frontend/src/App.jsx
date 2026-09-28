import React, { useState } from 'react';
import MovieGrid, { MOVIES } from './components/MovieGrid';
import PaymentModal from './components/PaymentModal';
import VideoPlayer from './components/VideoPlayer';
import { getContent } from './services/api';
import { Play, Info, Star, Clock, X } from 'lucide-react';

// Feature spotlight panel shown when a movie is selected from the grid
function FeaturePanel({ movie, onPlay, onClose }) {
  return (
    <div className="relative w-full overflow-hidden" style={{ height: '65vh' }}>
      {/* Background image with gradients */}
      <div className="absolute inset-0">
        <img
          src={movie.image}
          alt={movie.title}
          className="w-full h-full object-cover object-top"
          style={{ transition: 'opacity 0.5s ease' }}
        />
        <div className="absolute inset-0 bg-gradient-to-r from-black via-black/60 to-transparent" />
        <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-transparent to-black/40" />
      </div>

      {/* Close button */}
      <button
        onClick={onClose}
        className="absolute top-5 right-6 z-20 text-gray-400 hover:text-white bg-black/50 rounded-full p-1.5 transition"
      >
        <X size={18} />
      </button>

      {/* Content */}
      <div className="absolute bottom-[12%] left-12 max-w-2xl z-20">
        <div className="flex items-center gap-3 mb-3">
          <span className="text-xs font-semibold bg-[#E50914] px-2 py-0.5 rounded uppercase tracking-widest">Now Featured</span>
          <span className="text-xs text-gray-400">{movie.year}</span>
          <span className="text-xs border border-gray-500 text-gray-300 px-1.5 py-0.5 rounded">{movie.rating}</span>
          <span className="text-xs text-gray-400 flex items-center gap-1"><Clock size={11} />{movie.duration}</span>
        </div>

        <h2 className="text-5xl font-extrabold mb-3 drop-shadow-xl leading-tight">{movie.title}</h2>
        <p className="text-sm text-gray-300 mb-2 font-medium tracking-wide">{movie.genre}</p>
        <p className="text-base text-gray-200 mb-8 leading-relaxed max-w-xl">{movie.description}</p>

        <div className="flex items-center gap-4">
          <button
            onClick={() => onPlay(movie.id)}
            className="flex items-center gap-2 bg-white text-black px-8 py-3 rounded font-bold hover:bg-gray-200 transition shadow-xl text-sm"
          >
            <Play fill="black" size={18} />
            Play
          </button>
          <button className="flex items-center gap-2 bg-gray-600/70 hover:bg-gray-500/70 text-white px-6 py-3 rounded font-semibold transition text-sm backdrop-blur-sm border border-gray-500/30">
            <Info size={18} />
            More Info
          </button>
        </div>
      </div>
    </div>
  );
}

const HERO_MOVIE = {
  id: 'mov-hero',
  title: 'The Matrix Resurrections',
  genre: 'Action · Sci-Fi',
  year: 2023,
  rating: 'R',
  duration: '2h 28m',
  description: 'Return to a world of two realities: one, everyday life; the other, what lies behind it. To find out if his reality is a construct, to truly know himself, Mr. Anderson will have to follow the white rabbit once more.',
  image: 'https://images.unsplash.com/photo-1536440136628-849c177e76a1?w=1200&auto=format&fit=crop&q=80',
};

function App() {
  const [activeVideo, setActiveVideo] = useState(null);
  const [paymentTemplate, setPaymentTemplate] = useState(null);
  const [currentReceipt, setCurrentReceipt] = useState(null);
  const [selectedMovie, setSelectedMovie] = useState(null);

  const handleMovieSelect = (movie) => {
    // If same movie clicked again, deselect
    setSelectedMovie(prev => prev?.id === movie.id ? null : movie);
  };

  const handlePlayRequest = async (contentId) => {
    try {
      const result = await getContent(contentId);
      if (result.status === 'success') {
        setActiveVideo({ id: contentId, url: result.content_url });
      }
    } catch (err) {
      // 402 — extract the template and show payment modal
      if (err.response?.status === 402) {
        const detail = err.response.data?.detail;
        setPaymentTemplate({
          ...detail?.x402_payload_template,
          resource_identifier: contentId,
        });
      } else {
        alert('An error occurred: ' + (err.message || 'Unknown error'));
      }
    }
  };

  const handlePaymentSuccess = async (receiptId) => {
    setCurrentReceipt(receiptId);
    const targetId = paymentTemplate?.resource_identifier;
    setPaymentTemplate(null);

    if (targetId) {
      try {
        const result = await getContent(targetId, receiptId);
        if (result.status === 'success') {
          setSelectedMovie(null);
          setActiveVideo({ id: targetId, url: result.content_url });
        }
      } catch {
        alert('Failed to unlock content after payment.');
      }
    }
  };

  // Determine what to show in the feature area: selected grid movie or default hero
  const featureMovie = selectedMovie || HERO_MOVIE;
  const isCustomSelection = !!selectedMovie;

  return (
    <div className="min-h-screen bg-[#141414] text-white">
      {/* Sticky Header */}
      <header className="px-12 py-5 flex items-center justify-between sticky top-0 bg-gradient-to-b from-black/90 to-transparent z-30">
        <div className="text-[#E50914] font-black text-3xl tracking-tighter select-none">NEXUS</div>
        <nav className="flex gap-6 text-sm font-semibold">
          <span className="text-white cursor-pointer">Home</span>
          <span className="text-gray-400 hover:text-white cursor-pointer transition">Series</span>
          <span className="text-gray-400 hover:text-white cursor-pointer transition">Films</span>
          <span className="text-gray-400 hover:text-white cursor-pointer transition">New & Hot</span>
        </nav>
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-md bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-xs font-extrabold shadow-lg cursor-pointer border border-blue-400/40">
            W
          </div>
        </div>
      </header>

      {/* Feature Panel — always visible (hero default, or selected movie) */}
      {!activeVideo && (
        <div className="relative -mt-[72px]">
          {isCustomSelection ? (
            <FeaturePanel
              movie={featureMovie}
              onPlay={handlePlayRequest}
              onClose={() => setSelectedMovie(null)}
            />
          ) : (
            /* Default Hero */
            <div className="relative h-[65vh] w-full">
              <div className="absolute inset-0">
                <img
                  src={HERO_MOVIE.image}
                  alt={HERO_MOVIE.title}
                  className="w-full h-full object-cover"
                />
                <div className="absolute inset-0 bg-gradient-to-r from-black via-black/60 to-transparent" />
                <div className="absolute inset-0 bg-gradient-to-t from-[#141414] to-black/30" />
              </div>
              <div className="absolute bottom-[15%] left-12 max-w-xl z-20">
                <h1 className="text-5xl font-extrabold mb-4 drop-shadow-lg leading-tight">{HERO_MOVIE.title}</h1>
                <p className="text-base text-gray-200 mb-7 leading-relaxed">{HERO_MOVIE.description}</p>
                <div className="flex gap-4">
                  <button
                    onClick={() => handlePlayRequest(HERO_MOVIE.id)}
                    className="flex items-center gap-2 bg-white text-black px-8 py-3 rounded font-bold hover:bg-gray-200 transition shadow-xl text-sm"
                  >
                    <Play fill="black" size={18} />
                    Play
                  </button>
                  <button className="flex items-center gap-2 bg-gray-600/70 hover:bg-gray-500/70 text-white px-6 py-3 rounded font-semibold transition text-sm backdrop-blur-sm border border-gray-500/30">
                    <Info size={18} />
                    More Info
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Movie Catalog Grid */}
      <main className={`relative z-20 pb-24 ${!activeVideo ? '-mt-8' : 'mt-0'}`}>
        <MovieGrid
          selectedMovieId={selectedMovie?.id}
          onSelect={handleMovieSelect}
        />
      </main>

      {/* Payment Modal */}
      {paymentTemplate && (
        <PaymentModal
          template={paymentTemplate}
          onClose={() => setPaymentTemplate(null)}
          onSuccess={handlePaymentSuccess}
        />
      )}

      {/* Full-screen Video Player */}
      {activeVideo && (
        <VideoPlayer
          contentUrl={activeVideo.url}
          receiptId={currentReceipt}
          onBack={() => setActiveVideo(null)}
        />
      )}
    </div>
  );
}

export default App;
