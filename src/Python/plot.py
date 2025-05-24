import json
import matplotlib.pyplot as plt
import numpy as np

# Load data from separate JSON files
with open('keygen_performance.json', 'r') as f:
    keygen_data = json.load(f)
with open('sign_performance.json', 'r') as f:
    sign_data = json.load(f)
with open('combine_performance.json', 'r') as f:
    combine_data = json.load(f)
with open('verify_performance.json', 'r') as f:
    verify_data = json.load(f)

# Extract data and labels
sQ_values = [int(k) for k in keygen_data.keys()]
keygen_times = [np.mean(keygen_data[str(sQ)]['keygen_times']) for sQ in sQ_values]
sign_times = [np.mean(sign_data[str(sQ)]['sign_times']) for sQ in sQ_values]
combine_times = [np.mean(combine_data[str(sQ)]['combine_times']) for sQ in sQ_values]
verify_times = [np.mean(verify_data[str(sQ)]['verify_times']) for sQ in sQ_values]
keygen_errors = [np.std(keygen_data[str(sQ)]['keygen_times']) for sQ in sQ_values]
sign_errors = [np.std(sign_data[str(sQ)]['sign_times']) for sQ in sQ_values]
combine_errors = [np.std(combine_data[str(sQ)]['combine_times']) for sQ in sQ_values]
verify_errors = [np.std(verify_data[str(sQ)]['verify_times']) for sQ in sQ_values]
labels = [str(sQ) for sQ in sQ_values]

COLORS = {
    'keygen': '#800080',
    'signing': '#0000FF',
    'combining': '#008000',
    'verification': '#FF8C00',
    'light_blue': '#FA4616',
    'sky_blue': '#00BFFF'
}

# Function to save bar plot with error bars, values on top, and dashed grid lines
def save_plot(data, errors, ylabel, color, filename):
    fig, ax = plt.subplots(figsize=(8, 6))
    x = np.arange(len(sQ_values))
    bar_width = 0.5  # Adjusted bar width to match the earlier plot's spacing
    bars = ax.bar(x, data, bar_width, color=color, yerr=errors, capsize=5, ecolor='black')

    # Add values on top of the bars
    for bar in bars:
        yval = bar.get_height()
        # Format the value to 2 decimal places and position it slightly above the bar
        ax.text(bar.get_x() + bar.get_width()/2, yval + 0.05 * max(data), f'{yval:.2f}', 
                ha='center', va='bottom', fontsize=10)

    ax.set_xlabel('Number of Peers in Quorum (sQ)')
    ax.set_ylabel(ylabel)
    ax.set_xticks(x)
    ax.set_xticklabels(labels)

    # Customize grid lines to be lighter and dashed
    ax.grid(True, linestyle='--', alpha=0.5)  # Dashed lines with reduced opacity

    plt.tight_layout()
    plt.savefig(filename)
    plt.close()

# Save plots
save_plot(keygen_times, keygen_errors, 'Time (s)', COLORS['keygen'], '5keygen_time_final.png')
save_plot(sign_times, sign_errors, 'Time (ms)', COLORS['light_blue'], '5sign_time_final.png')
save_plot(combine_times, combine_errors, 'Time (ms)', COLORS['sky_blue'], '5combine_time_final.png')
save_plot(verify_times, verify_errors, 'Time (ms)', COLORS['combining'], '5verify_time_l.png')