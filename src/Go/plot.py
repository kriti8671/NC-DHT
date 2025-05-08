import json
import matplotlib.pyplot as plt

# Load the benchmark results
with open("benchmark_results.json", "r") as f:
    data = json.load(f)

# Extract relevant values
num_quorums = [entry["num_quorums"] for entry in data]
encoding_times = [entry["encoding_time_ms"] for entry in data]
decoding_times = [entry["decoding_time_ms"] for entry in data]

# Plotting
plt.figure(figsize=(10, 6))
plt.plot(num_quorums, encoding_times, marker='o', color='green', label='Encoding Time (ms)')
plt.plot(num_quorums, decoding_times, marker='x', color='orange', label='Decoding Time (ms)')

plt.xlabel("Number of Quorums")
plt.ylabel("Time (ms)")
plt.title("Encoding vs Decoding Time vs Number of Quorums")
plt.grid(True)
plt.legend()
plt.tight_layout()
plt.savefig("benchmark_plot.png")  # Optional: save to file
plt.show()
