#pragma once

#include <opencv2/opencv.hpp>
#include <string>

/**
 * Utility functions for drawing bounding boxes and labels
 */
namespace DrawingUtils {
    // Constants for label positioning
    constexpr int LABEL_TEXT_BASELINE_OFFSET = 5;  // Offset from bbox top/bottom edge
    constexpr int LABEL_BACKGROUND_PADDING_TOP = 2; // Padding above text
    constexpr int LABEL_BACKGROUND_PADDING_BOTTOM = 2; // Padding below text
    
    /**
     * Draw a label with background for a detection bounding box
     * Automatically positions the label at the top or bottom of the bbox to avoid cutoff
     * 
     * @param frame Frame to draw on (modified in place)
     * @param label Text to display
     * @param bbox Bounding box rectangle
     * @param color Color for the background rectangle
     * @param font_face OpenCV font face (default: cv::FONT_HERSHEY_SIMPLEX)
     * @param font_scale Font scale (default: 0.5)
     * @param font_thickness Font thickness (default: 1)
     */
    inline void drawBoundingBoxLabel(cv::Mat& frame,
                                     const std::string& label,
                                     const cv::Rect& bbox,
                                     const cv::Scalar& color,
                                     int font_face = cv::FONT_HERSHEY_SIMPLEX,
                                     double font_scale = 0.5,
                                     int font_thickness = 1) {
        int baseline;
        cv::Size text_size = cv::getTextSize(label, font_face, font_scale, font_thickness, &baseline);
        
        // Calculate the top of the label background if positioned above bbox
        int top_rect_y = bbox.y - LABEL_TEXT_BASELINE_OFFSET - text_size.height - LABEL_BACKGROUND_PADDING_TOP;
        
        // Position label at top or bottom based on whether it would be cut off
        cv::Point text_origin;
        if (top_rect_y < 0) {
            // Label would be cut off at top - position at bottom of bbox instead
            text_origin = cv::Point(bbox.x, bbox.y + bbox.height + text_size.height + LABEL_TEXT_BASELINE_OFFSET);
        } else {
            // Normal position at top of bbox
            text_origin = cv::Point(bbox.x, bbox.y - LABEL_TEXT_BASELINE_OFFSET);
        }
        
        // Draw label background rectangle
        cv::rectangle(frame,
                     cv::Point(text_origin.x, text_origin.y - text_size.height - LABEL_BACKGROUND_PADDING_TOP),
                     cv::Point(text_origin.x + text_size.width, text_origin.y + LABEL_BACKGROUND_PADDING_BOTTOM),
                     color, cv::FILLED);
        
        // Draw label text
        cv::putText(frame, label, text_origin, 
                   font_face, font_scale, cv::Scalar(0, 0, 0), font_thickness);
    }
}
