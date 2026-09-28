import React, { useState } from 'react';
import { verifyPayment } from '../services/api';
import { X, Loader2, CheckCircle, Receipt as ReceiptIcon } from 'lucide-react';

export default function PaymentModal({ template, onClose, onSuccess }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [selectedNetwork, setSelectedNetwork] = useState(null);

  const handlePay = async (network, currency, amountScale) => {
    setSelectedNetwork(network);
    setLoading(true);
    setError(null);
    try {
      // Simulate wallet signing delay
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      const payload = {
        payment_id: `PAY-${Date.now()}`,
        transaction_id: `TX-${Math.random().toString(36).substring(2, 15).toUpperCase()}`,
        amount: template.amount * amountScale,
        currency: currency,
        scheme: "x402",
        network: network,
        recipient_address: template.recipient_address,
        resource_identifier: template.resource_identifier,
        nonce: Date.now().toString(),
        signature: "mock_signature" // Bypasses algosdk verify in our mock backend
      };

      const result = await verifyPayment(payload);
      
      if (result.status === "success") {
        setSuccess(true);
        setTimeout(() => {
            onSuccess(result.receipt_id);
        }, 1500);
      }
    } catch (err) {
      setError(err.response?.data?.detail || err.message || "Payment failed");
      setSelectedNetwork(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
      <div className="bg-[#181818] w-full max-w-md rounded-xl shadow-2xl overflow-hidden border border-gray-800">
        <div className="p-6 relative">
          <button onClick={onClose} className="absolute top-4 right-4 text-gray-400 hover:text-white transition">
            <X size={20} />
          </button>
          
          <h2 className="text-2xl font-bold mb-2">Unlock Content</h2>
          <p className="text-gray-400 text-sm mb-6">402 Payment Required for {template.resource_identifier}</p>

          {!success && !loading && (
             <div className="bg-black/50 rounded-lg p-4 mb-6 border border-gray-800 flex flex-col gap-3">
                <h3 className="text-gray-300 text-sm font-semibold mb-2">Select Payment Method</h3>
                
                {/* Ethereum Option */}
                <button 
                   onClick={() => handlePay("Ethereum", "ETH", 0.0005)}
                   disabled={loading}
                   className="w-full flex items-center justify-between p-3 rounded border border-gray-700 hover:border-[#627EEA] hover:bg-gray-800 transition disabled:opacity-50"
                >
                   <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-[#627EEA]/20 text-[#627EEA] flex items-center justify-center text-xs font-bold">ETH</div>
                      <span className="font-semibold text-gray-200">Ethereum</span>
                   </div>
                   <span className="text-sm">{(template.amount * 0.0005).toFixed(4)} ETH</span>
                </button>

                {/* Solana Option */}
                <button 
                   onClick={() => handlePay("Solana", "SOL", 0.01)}
                   disabled={loading}
                   className="w-full flex items-center justify-between p-3 rounded border border-gray-700 hover:border-[#14F195] hover:bg-gray-800 transition disabled:opacity-50"
                >
                   <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-[#14F195]/20 text-[#14F195] flex items-center justify-center text-xs font-bold">SOL</div>
                      <span className="font-semibold text-gray-200">Solana</span>
                   </div>
                   <span className="text-sm">{(template.amount * 0.01).toFixed(2)} SOL</span>
                </button>

                {/* Algorand Option */}
                <button 
                   onClick={() => handlePay("Algorand", "ALGO", 1)}
                   disabled={loading}
                   className="w-full flex items-center justify-between p-3 rounded border border-gray-700 hover:border-white hover:bg-gray-800 transition disabled:opacity-50"
                >
                   <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-white text-black flex items-center justify-center text-xs font-bold">ALG</div>
                      <span className="font-semibold text-gray-200">Algorand</span>
                   </div>
                   <span className="text-sm">{template.amount.toFixed(2)} ALGO</span>
                </button>
             </div>
          )}

          {error && (
            <div className="bg-red-500/10 border border-red-500/50 text-red-500 text-sm p-3 rounded-md mb-6">
              {error}
            </div>
          )}

          {success ? (
            <div className="flex flex-col items-center justify-center py-6 text-green-400">
              <CheckCircle size={48} className="mb-4" />
              <p className="font-bold">Payment Settled</p>
              <p className="text-sm text-gray-400 mt-2">Generating Receipt...</p>
            </div>
          ) : loading ? (
             <div className="w-full bg-[#E50914] text-white font-bold py-3 px-4 rounded flex items-center justify-center">
                 <Loader2 className="animate-spin mr-2" size={20} />
                 Awaiting {selectedNetwork} Signature...
             </div>
          ) : null}
        </div>
        <div className="bg-[#141414] p-4 text-xs text-gray-500 flex items-center justify-center border-t border-gray-800">
          <ReceiptIcon size={14} className="mr-2" />
          Powered by x402 Non-Custodial Mesh
        </div>
      </div>
    </div>
  );
}
