import React from 'react';
import { ArrowLeft, Download } from 'lucide-react';
import { getReceiptUrl } from '../services/api';

export default function VideoPlayer({ contentUrl, receiptId, onBack }) {
  return (
    <div className="fixed inset-0 bg-black z-40 flex flex-col">
      <div className="flex justify-between items-center p-6 bg-gradient-to-b from-black/80 to-transparent absolute top-0 w-full z-50">
        <button onClick={onBack} className="text-white hover:text-gray-300 transition flex items-center gap-2">
          <ArrowLeft size={24} />
          <span className="font-bold">Back to Browse</span>
        </button>
        <a 
          href={getReceiptUrl(receiptId)} 
          target="_blank" 
          rel="noopener noreferrer"
          className="flex items-center gap-2 bg-white/10 hover:bg-white/20 backdrop-blur-md border border-white/20 px-4 py-2 rounded text-sm text-white transition"
        >
          <Download size={16} />
          Download Receipt
        </a>
      </div>
      <div className="flex-1 flex items-center justify-center bg-black">
        <video 
          controls 
          autoPlay 
          className="w-full max-h-screen object-contain"
          src={contentUrl}
        />
      </div>
    </div>
  );
}
