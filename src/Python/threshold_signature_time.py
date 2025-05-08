import time
import matplotlib.pyplot as plt
import random
import uuid
import logging
import sys
import os
from base64 import b64encode

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
PEER_COUNTS = [ 10, 50,100,500]
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

def plot_performance():
    sign_data = []
    combine_data = []
    verify_data = []
    keygen_data = []
    labels = [str(sQ) for sQ in PEER_COUNTS]
    
    for sQ in PEER_COUNTS:
        perf = simulate(sQ, NUM_RUNS)
        if perf:
            keygen_data.append(perf['keygen_times'])
            sign_data.append(perf['sign_times'])
            combine_data.append(perf['combine_times'])
            verify_data.append(perf['verify_times'])
    # print("keygen_data: \n",keygen_data)
    # print("keygen_data: \n",combine_data)
    # Calculate and print average times for each sQ
    print("\nAverage Times (in seconds) for each Number of Peers (sQ):")
    for i, sQ in enumerate(PEER_COUNTS):
        avg_keygen = sum(keygen_data[i]) / len(keygen_data[i]) if keygen_data[i] else 0
        avg_sign = sum(sign_data[i]) / len(sign_data[i]) if sign_data[i] else 0
        avg_combine = sum(combine_data[i]) / len(combine_data[i]) if combine_data[i] else 0
        avg_verify = sum(verify_data[i]) / len(verify_data[i]) if verify_data[i] else 0
        print(f"sQ = {sQ}:")
        print(f"  Average Key Generation Time: {avg_keygen:.6f} s")
        print(f"  Average Signing Time: {avg_sign:.6f} s")
        print(f"  Average Combining Time: {avg_combine:.8f} s")
        print(f"  Average Verification Time: {avg_verify:.6f} s")
    fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(12, 10))
    
    ax1.boxplot(keygen_data, labels=labels, patch_artist=True,
                boxprops=dict(facecolor=COLORS['keygen'], color=COLORS['keygen']),
                whiskerprops=dict(color=COLORS['keygen']),
                capprops=dict(color=COLORS['keygen']),
                medianprops=dict(color='black'))
    ax1.set_title('Key Generation Time vs. Number of Peers in Quorum')
    ax1.set_xlabel('Number of Peers in Quorum (sQ)')
    ax1.set_ylabel('Time (seconds)')
    ax1.grid(True)
    
    ax2.boxplot(sign_data, labels=labels, patch_artist=True,
                boxprops=dict(facecolor=COLORS['signing'], color=COLORS['signing']),
                whiskerprops=dict(color=COLORS['signing']),
                capprops=dict(color=COLORS['signing']),
                medianprops=dict(color='black'))
    ax2.set_title('Signing Time (t+1 peers) vs. Number of Peers in Quorum')
    ax2.set_xlabel('Number of Peers in Quorum (sQ)')
    ax2.set_ylabel('Time (seconds)')
    ax2.grid(True)
    
    ax3.boxplot(combine_data, labels=labels, patch_artist=True,
                boxprops=dict(facecolor=COLORS['combining'], color=COLORS['combining']),
                whiskerprops=dict(color=COLORS['combining']),
                capprops=dict(color=COLORS['combining']),
                medianprops=dict(color='black'))
    ax3.set_title('Combining Time vs. Number of Peers in Quorum')
    ax3.set_xlabel('Number of Peers in Quorum (sQ)')
    ax3.set_ylabel('Time (seconds)')
    ax3.grid(True)
    
    ax4.boxplot(verify_data, labels=labels, patch_artist=True,
                boxprops=dict(facecolor=COLORS['verification'], color=COLORS['verification']),
                whiskerprops=dict(color=COLORS['verification']),
                capprops=dict(color=COLORS['verification']),
                medianprops=dict(color='black'))
    ax4.set_title('Verification Time vs. Number of Peers in Quorum')
    ax4.set_xlabel('Number of Peers in Quorum (sQ)')
    ax4.set_ylabel('Time (seconds)')
    ax4.grid(True)
    
    plt.tight_layout()
    plt.savefig('keygen500_performance_boxplot_Performance.png')
    plt.close()

if __name__ == "__main__":
    plot_performance()