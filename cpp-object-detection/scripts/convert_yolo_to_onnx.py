#!/usr/bin/env python3
"""
YOLO Model Converter - PyTorch (.pt) to ONNX format
Converts YOLOv9, YOLOv10, and YOLOv11 models to ONNX format for use with C++ applications.

Usage:
    python convert_yolo_to_onnx.py --version 11 --model-size n
    python convert_yolo_to_onnx.py --version 10 --model-size s --output-dir ./models
    python convert_yolo_to_onnx.py --pt-file yolo11n.pt --output yolo11n.onnx

Requirements:
    pip install ultralytics onnx
"""

import argparse
import os
import sys
from pathlib import Path

def main():
    parser = argparse.ArgumentParser(
        description='Convert YOLO models from PyTorch (.pt) to ONNX format',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Convert YOLOv11n model (downloads automatically if needed)
  python convert_yolo_to_onnx.py --version 11 --model-size n

  # Convert YOLOv10s model with custom output directory
  python convert_yolo_to_onnx.py --version 10 --model-size s --output-dir ./custom_models

  # Convert existing .pt file to ONNX
  python convert_yolo_to_onnx.py --pt-file yolo11n.pt --output yolo11n.onnx

Supported YOLO versions: 9, 10, 11
Supported model sizes: n (nano), s (small), m (medium), l (large), x (xlarge)
        """
    )
    
    # Model selection arguments
    model_group = parser.add_mutually_exclusive_group(required=True)
    model_group.add_argument(
        '--version',
        type=int,
        choices=[9, 10, 11],
        help='YOLO version to convert (9, 10, or 11). Requires --model-size.'
    )
    model_group.add_argument(
        '--pt-file',
        type=str,
        help='Path to existing .pt file to convert'
    )
    
    parser.add_argument(
        '--model-size',
        type=str,
        choices=['n', 's', 'm', 'l', 'x'],
        help='Model size: n (nano), s (small), m (medium), l (large), x (xlarge). Required when using --version.'
    )
    
    parser.add_argument(
        '--output-dir',
        type=str,
        default='./models',
        help='Output directory for ONNX file (default: ./models)'
    )
    
    parser.add_argument(
        '--output',
        type=str,
        help='Custom output filename (overrides automatic naming)'
    )
    
    parser.add_argument(
        '--imgsz',
        type=int,
        default=640,
        help='Image size for ONNX export (default: 640)'
    )
    
    parser.add_argument(
        '--simplify',
        action='store_true',
        default=True,
        help='Simplify ONNX model using onnxsim (default: True)'
    )
    
    parser.add_argument(
        '--opset',
        type=int,
        default=12,
        help='ONNX opset version (default: 12)'
    )
    
    parser.add_argument(
        '--verbose',
        action='store_true',
        help='Verbose output'
    )
    
    args = parser.parse_args()
    
    # Validate arguments
    if args.version and not args.model_size:
        parser.error('--model-size is required when using --version')
    
    # Import ultralytics (check if installed)
    try:
        from ultralytics import YOLO
        import onnx
    except ImportError as e:
        print(f"Error: Required Python packages not installed: {e}", file=sys.stderr)
        print("\nPlease install required packages:", file=sys.stderr)
        print("  pip install ultralytics onnx", file=sys.stderr)
        sys.exit(1)
    
    # Determine model source
    if args.pt_file:
        pt_path = Path(args.pt_file)
        if not pt_path.exists():
            print(f"Error: .pt file not found: {args.pt_file}", file=sys.stderr)
            sys.exit(1)
        model_name = pt_path.stem
        print(f"Converting existing model: {args.pt_file}")
    else:
        # Construct model name based on version and size
        model_name = f"yolo{args.version}{args.model_size}"
        pt_path = None
        print(f"Converting {model_name} model (will download if needed)")
    
    # Setup output
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    if args.output:
        output_file = Path(args.output)
        if not output_file.is_absolute():
            output_file = output_dir / output_file
    else:
        output_file = output_dir / f"{model_name}.onnx"
    
    print(f"Output will be saved to: {output_file}")
    
    # Load model
    try:
        if pt_path:
            print(f"Loading model from {pt_path}...")
            model = YOLO(str(pt_path))
        else:
            print(f"Loading {model_name} model...")
            # This will download the model if not already cached
            model = YOLO(f"{model_name}.pt")
        
        print("Model loaded successfully!")
        
    except Exception as e:
        print(f"Error loading model: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Export to ONNX
    try:
        print(f"\nExporting to ONNX format...")
        print(f"  Image size: {args.imgsz}")
        print(f"  ONNX opset: {args.opset}")
        print(f"  Simplify: {args.simplify}")
        
        # Export using ultralytics built-in export
        export_path = model.export(
            format='onnx',
            imgsz=args.imgsz,
            opset=args.opset,
            simplify=args.simplify,
            dynamic=False,  # Static input size for better performance
        )
        
        print(f"\nExport completed: {export_path}")
        
        # Move to desired output location if different
        export_path = Path(export_path)
        if export_path != output_file:
            import shutil
            shutil.move(str(export_path), str(output_file))
            print(f"Moved to: {output_file}")
        
    except Exception as e:
        print(f"Error during ONNX export: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Validate ONNX model
    try:
        print("\nValidating ONNX model...")
        onnx_model = onnx.load(str(output_file))
        onnx.checker.check_model(onnx_model)
        print("✓ ONNX model is valid!")
        
        # Print model info
        file_size_mb = output_file.stat().st_size / (1024 * 1024)
        print(f"\nModel Information:")
        print(f"  File: {output_file}")
        print(f"  Size: {file_size_mb:.2f} MB")
        print(f"  ONNX version: {onnx_model.opset_import[0].version}")
        
        # Print input/output shapes
        if args.verbose:
            print(f"\n  Inputs:")
            for input_tensor in onnx_model.graph.input:
                shape = [dim.dim_value for dim in input_tensor.type.tensor_type.shape.dim]
                print(f"    {input_tensor.name}: {shape}")
            
            print(f"\n  Outputs:")
            for output_tensor in onnx_model.graph.output:
                shape = [dim.dim_value for dim in output_tensor.type.tensor_type.shape.dim]
                print(f"    {output_tensor.name}: {shape}")
        
        print("\n✓ Conversion successful!")
        
    except Exception as e:
        print(f"Warning: ONNX validation failed: {e}", file=sys.stderr)
        print("The model file was created but may have issues.", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
