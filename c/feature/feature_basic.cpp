//
// Created by shuli on 2024/3/5.
//

#include "feature_basic.h"
#include "feature_functions.h"

int speakerlab::FrameExtractionOptions::compute_window_shift() {
    return static_cast<int>(sample_freq * 0.001 * frame_shift_ms);
}

int speakerlab::FrameExtractionOptions::compute_window_size() {
    return static_cast<int>(sample_freq * 0.001 * frame_length_ms);
}

int speakerlab::FrameExtractionOptions::padded_window_size() {
    int window_size = compute_window_size();
    if(round_to_power_of_two) {
        return round_up_to_nearest_power_of_two(window_size);
    }
    else {
        return window_size;
    }
}
