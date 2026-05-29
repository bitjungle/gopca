"""Create a reference-backed electrode-position map for the EEG Eye State dataset.

The channel set matches the 14 Emotiv EPOC/EPOC X sensor locations.
Coordinates are taken from MNE-Python's built-in ``standard_1020`` montage.

Output:
    eeg_electrode_map.png
"""

from pathlib import Path

import matplotlib.pyplot as plt
import mne

CHANNELS = [
    "AF3", "F7", "F3", "FC5", "T7", "P7", "O1",
    "O2", "P8", "T8", "FC6", "F4", "F8", "AF4",
]

OUTPUT = Path("eeg_electrode_map.png")


def main() -> None:
    # MNE's standard_1020 montage provides canonical EEG sensor coordinates.
    montage = mne.channels.make_standard_montage("standard_1020")

    missing = sorted(set(CHANNELS) - set(montage.ch_names))
    if missing:
        raise ValueError(f"Channels missing from MNE standard_1020 montage: {missing}")

    info = mne.create_info(ch_names=CHANNELS, sfreq=128, ch_types="eeg")
    info.set_montage(montage)

    fig = plt.figure(figsize=(7, 7))
    ax = fig.add_subplot(111)

    # Plot only the EEG sensor locations using MNE's topomap projection.
    # A fixed head sphere avoids depending on extra fiducial channels such as Fpz/Oz.
    mne.viz.plot_sensors(
        info,
        kind="topomap",
        show_names=True,
        axes=ax,
        show=False,
        sphere=(0.0, 0.0, 0.0, 0.095),
        pointsize=70,
        linewidth=1.5,
    )

    ax.set_title("EEG Eye State Dataset: 14 Emotiv EPOC Sensor Positions", pad=18)

    # Add a small note directly in the figure for provenance.
    fig.text(
        0.5,
        0.025,
        "Channel set: Emotiv EPOC/EPOC X. Coordinates: MNE standard_1020 montage.",
        ha="center",
        va="bottom",
        fontsize=9,
    )

    fig.savefig(OUTPUT, dpi=300, bbox_inches="tight")
    plt.close(fig)
    print(f"Saved {OUTPUT.resolve()}")


if __name__ == "__main__":
    main()
