import time
import matplotlib.pyplot as plt
import random
import uuid
import logging
import sys
import os
from base64 import b64encode
import json

# Adjust path to include repository files
repo_path = os.path.abspath("./threshold-signature-demo")
if repo_path not in sys.path:
    sys.path.append(repo_path)

from threshold_signature import ThresholdSignature
from sign import verify_signature, hash_to_int, sign
from ec_point_operation import scalar_multiply, curve
from polynomial import Polynomial

# Set up logging (commented out but retained for future use)
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Configuration for peer counts and number of runs
PEER_COUNTS = [5, 10, 50, 100, 150]
# PEER_COUNTS=[5,10,20,40,80,160]
NUM_RUNS = 5

# Color configuration for plots
COLORS = {
    'keygen': '#800080',
    'signing': '#0000FF',
    'combining': '#008000',
    'verification': '#FF8C00'
}

class Peer:
    def __init__(self, id, quorum_id, private_key_share, public_key_share):
        self.id = id
        self.quorum_id = quorum_id
        self.private_key_share = private_key_share
        self.public_key_share = public_key_share
    
    def process_request(self, message):
        r, s = sign(self.private_key_share, message.encode())
        serialized_sig = r.to_bytes(32, 'big') + s.to_bytes(32, 'big')
        return (r, s)

class Quorum:
    def __init__(self, id, sQ, tQ):
        self.id = id
        self.sQ = sQ
        self.tQ = tQ
        start_time = time.perf_counter()
        self.ts_instance = ThresholdSignature(group_size=sQ, threshold=tQ + 1)
        self.keygen_time = time.perf_counter() - start_time
        self.public_key = self.ts_instance.public_key
        self.private_key_shares = self.ts_instance.shares
        self.public_key_shares = [scalar_multiply(s, curve.g) for s in self.private_key_shares]
        self.peers = [Peer(i, id, self.private_key_shares[i], self.public_key_shares[i]) for i in range(sQ)]

    def respond(self, message, num_signers):
        available_peers = list(range(self.sQ))
        signers = random.sample(available_peers, num_signers)
        signatures = []
        start_time = time.perf_counter()
        for peer_id in signers:
            peer = self.peers[peer_id]
            sig = peer.process_request(message)
            signatures.append(sig)
        sign_time = time.perf_counter() - start_time
        valid = all(sig is not None for sig in signatures)
        return signatures, valid, signers, sign_time

    def combine_shares(self, signatures, share_ids, message):
        start_time = time.perf_counter()
        points = [(i + 1, sig[1]) for i, sig in zip(share_ids, signatures)]
        s = Polynomial.interpolate_evaluate(points, 0) % curve.n
        r = signatures[0][0]
        combine_time = time.perf_counter() - start_time
        return (r, s), combine_time

class Initiator:
    def __init__(self, quorum, sQ, tQ):
        self.quorum = quorum
        self.sQ = sQ
        self.tQ = tQ
        self.id = str(uuid.uuid4())
        self.peer = quorum.peers[0]
        self.performance = {'keygen_times': [], 'sign_times': [], 'combine_times': [], 'verify_times': []}
    
    def lookup(self):
        message = f"REQUEST|{self.id}|{time.time()}"
        self.performance['keygen_times'].append(self.quorum.keygen_time)
        signatures, valid, signers, sign_time = self.quorum.respond(message, num_signers=self.tQ + 1)
        self.performance['sign_times'].append(sign_time)
        
        if not valid:
            return False
        
        valid_shares = signatures
        combined_signature, combine_time = self.quorum.combine_shares(valid_shares, signers, message)
        self.performance['combine_times'].append(combine_time)
        verify_start = time.perf_counter()
        digest = hash_to_int(message.encode())
        verify_signature(self.quorum.public_key, message.encode(), combined_signature)
        verify_time = time.perf_counter() - verify_start
        self.performance['verify_times'].append(verify_time)
        
        return True

def simulate(sQ, num_runs):
    tQ = sQ // 3
    performance = {'keygen_times': [], 'sign_times': [], 'combine_times': [], 'verify_times': []}
    
    for _ in range(num_runs):
        quorum = Quorum("0", sQ, tQ)
        initiator = Initiator(quorum, sQ, tQ)
        success = initiator.lookup()
        if success:
            performance['keygen_times'].extend(initiator.performance['keygen_times'])
            performance['sign_times'].extend(initiator.performance['sign_times'])
            performance['combine_times'].extend(initiator.performance['combine_times'])
            performance['verify_times'].extend(initiator.performance['verify_times'])
    
    return performance

def save_performance_data():
    performance_data = {}
    for sQ in PEER_COUNTS:
        perf = simulate(sQ, NUM_RUNS)
        if perf:
            performance_data[str(sQ)] = {
                'keygen_times': [x * 1 for x in perf['keygen_times']],  # Keep in seconds
                'sign_times': [x * 1e3 for x in perf['sign_times']],    # Convert to ms
                'combine_times': [x * 1e3 for x in perf['combine_times']],  # Convert to ms
                'verify_times': [x * 1e3 for x in perf['verify_times']]   # Convert to ms
            }
    
    # Save to separate JSON files
    with open('keygen_performance.json', 'w') as f:
        json.dump({str(sQ): {'keygen_times': v['keygen_times']} for sQ, v in performance_data.items()}, f, indent=4)
    with open('sign_performance.json', 'w') as f:
        json.dump({str(sQ): {'sign_times': v['sign_times']} for sQ, v in performance_data.items()}, f, indent=4)
    with open('combine_performance.json', 'w') as f:
        json.dump({str(sQ): {'combine_times': v['combine_times']} for sQ, v in performance_data.items()}, f, indent=4)
    with open('verify_performance.json', 'w') as f:
        json.dump({str(sQ): {'verify_times': v['verify_times']} for sQ, v in performance_data.items()}, f, indent=4)

    # Print average in ms
    print("\nAverage Times (in milliseconds) for each Number of Peers (sQ):")
    for i, sQ in enumerate(PEER_COUNTS):
        avg_keygen = sum(performance_data[str(sQ)]['keygen_times']) / len(performance_data[str(sQ)]['keygen_times']) if performance_data[str(sQ)]['keygen_times'] else 0
        avg_sign = sum(performance_data[str(sQ)]['sign_times']) / len(performance_data[str(sQ)]['sign_times']) if performance_data[str(sQ)]['sign_times'] else 0
        avg_combine = sum(performance_data[str(sQ)]['combine_times']) / len(performance_data[str(sQ)]['combine_times']) if performance_data[str(sQ)]['combine_times'] else 0
        avg_verify = sum(performance_data[str(sQ)]['verify_times']) / len(performance_data[str(sQ)]['verify_times']) if performance_data[str(sQ)]['verify_times'] else 0
        print(f"sQ = {sQ}:")
        print(f"  Average Key Generation Time: {avg_keygen:.5f} s")
        print(f"  Average Signing Time: {avg_sign:.5f} ms")
        print(f"  Average Combining Time: {avg_combine:.5f} ms")
        print(f"  Average Verification Time: {avg_verify:.5f} ms")

if __name__ == "__main__":
    save_performance_data()