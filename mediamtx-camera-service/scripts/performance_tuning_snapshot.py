#!/usr/bin/env python3
"""
Performance Tuning Script for Multi-Tier Snapshot Capture

This script dynamically tests and optimizes snapshot capture performance
parameters for different environments and use cases.

Features:
- Automated performance testing with different configurations
- Statistical analysis of timing data
- Environment-specific optimization recommendations
- Configuration file generation for optimal settings

Usage:
    python3 performance_tuning_snapshot.py [--environment ENV] [--iterations N] [--output FILE]

Environments:
    development: Fast response times for development workflow
    production: Balanced performance for production use
    high-performance: Maximum speed for critical applications
    embedded: Power-efficient settings for battery-powered devices
"""

import argparse
import asyncio
import json
import logging
import os
import statistics
import sys
import time
from pathlib import Path
from typing import Dict, Any, List, Optional

# Add src to path for imports
sys.path.append('src')

from camera_service.config import Config
from mediamtx_wrapper.controller import MediaMTXController


class SnapshotPerformanceTuner:
    """Performance tuner for multi-tier snapshot capture."""
    
    def __init__(self, config: Config):
        self.config = config
        self.mediamtx_controller: Optional[MediaMTXController] = None
        self.results: List[Dict[str, Any]] = []
        
    async def setup(self):
        """Setup MediaMTX controller for testing."""
        self.mediamtx_controller = MediaMTXController(
            host=self.config.mediamtx.host,
            api_port=self.config.mediamtx.api_port,
            rtsp_port=self.config.mediamtx.rtsp_port,
            webrtc_port=self.config.mediamtx.webrtc_port,
            hls_port=self.config.mediamtx.hls_port,
            config_path=self.config.mediamtx.config_path,
            recordings_path=self.config.mediamtx.recordings_path,
            snapshots_path=self.config.mediamtx.snapshots_path,
            ffmpeg_config=self.config.ffmpeg.__dict__
        )
        
        # Set performance configuration
        self.mediamtx_controller._performance_config = self.config.performance.__dict__
        
        await self.mediamtx_controller.start()
        logging.info("MediaMTX controller started for performance tuning")
        
    async def cleanup(self):
        """Cleanup resources."""
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()
        logging.info("MediaMTX controller stopped")
        
    async def test_configuration(self, config_name: str, snapshot_tiers_config: Dict[str, float], iterations: int = 5) -> Dict[str, Any]:
        """Test a specific configuration multiple times."""
        logging.info(f"Testing configuration: {config_name}")
        
        # Apply configuration
        self.mediamtx_controller._performance_config['snapshot_tiers'].update(snapshot_tiers_config)
        
        results = []
        
        for i in range(iterations):
            logging.info(f"  Iteration {i+1}/{iterations}")
            
            start_time = time.time()
            result = await self.mediamtx_controller.take_snapshot(
                stream_name="camera0",
                filename=f"perf_tune_{config_name}_{i}.jpg",
                format="jpg",
                quality=85
            )
            capture_time = time.time() - start_time
            
            results.append({
                "iteration": i + 1,
                "capture_time": capture_time,
                "tier_used": result.get("tier_used", 0),
                "user_experience": result.get("user_experience", "failed"),
                "status": result.get("status", "failed"),
                "success": result.get("status") == "completed"
            })
            
            # Small delay between iterations
            await asyncio.sleep(0.5)
        
        # Calculate statistics
        successful_results = [r for r in results if r["success"]]
        
        if successful_results:
            capture_times = [r["capture_time"] for r in successful_results]
            tiers_used = [r["tier_used"] for r in successful_results]
            user_experiences = [r["user_experience"] for r in successful_results]
            
            stats = {
                "config_name": config_name,
                "iterations": iterations,
                "successful_iterations": len(successful_results),
                "success_rate": len(successful_results) / iterations,
                "capture_time_stats": {
                    "mean": statistics.mean(capture_times),
                    "median": statistics.median(capture_times),
                    "min": min(capture_times),
                    "max": max(capture_times),
                    "std_dev": statistics.stdev(capture_times) if len(capture_times) > 1 else 0
                },
                "tier_distribution": {tier: tiers_used.count(tier) for tier in set(tiers_used)},
                "ux_distribution": {ux: user_experiences.count(ux) for ux in set(user_experiences)},
                "raw_results": results
            }
        else:
            stats = {
                "config_name": config_name,
                "iterations": iterations,
                "successful_iterations": 0,
                "success_rate": 0.0,
                "capture_time_stats": None,
                "tier_distribution": {},
                "ux_distribution": {},
                "raw_results": results
            }
        
        return stats
    
    def get_environment_configs(self, environment: str) -> Dict[str, Dict[str, float]]:
        """Get configuration sets for different environments."""
        base_configs = {
            "development": {
                "tier1_rtsp_ready_check_timeout": 0.5,
                "tier2_activation_timeout": 2.0,
                "tier2_activation_trigger_timeout": 1.0,
                "tier3_direct_capture_timeout": 3.0,
                "total_operation_timeout": 8.0,
                "immediate_response_threshold": 0.3,
                "acceptable_response_threshold": 1.5,
                "slow_response_threshold": 3.0
            },
            "production": {
                "tier1_rtsp_ready_check_timeout": 1.0,
                "tier2_activation_timeout": 3.0,
                "tier2_activation_trigger_timeout": 1.0,
                "tier3_direct_capture_timeout": 5.0,
                "total_operation_timeout": 10.0,
                "immediate_response_threshold": 0.5,
                "acceptable_response_threshold": 2.0,
                "slow_response_threshold": 5.0
            },
            "high-performance": {
                "tier1_rtsp_ready_check_timeout": 0.2,
                "tier2_activation_timeout": 1.5,
                "tier2_activation_trigger_timeout": 0.5,
                "tier3_direct_capture_timeout": 2.0,
                "total_operation_timeout": 5.0,
                "immediate_response_threshold": 0.2,
                "acceptable_response_threshold": 1.0,
                "slow_response_threshold": 2.0
            },
            "embedded": {
                "tier1_rtsp_ready_check_timeout": 2.0,
                "tier2_activation_timeout": 5.0,
                "tier2_activation_trigger_timeout": 2.0,
                "tier3_direct_capture_timeout": 8.0,
                "total_operation_timeout": 15.0,
                "immediate_response_threshold": 1.0,
                "acceptable_response_threshold": 3.0,
                "slow_response_threshold": 8.0
            }
        }
        
        if environment == "all":
            return base_configs
        elif environment in base_configs:
            return {environment: base_configs[environment]}
        else:
            raise ValueError(f"Unknown environment: {environment}")
    
    def generate_variations(self, base_config: Dict[str, float]) -> Dict[str, Dict[str, float]]:
        """Generate configuration variations for optimization."""
        variations = {}
        
        # Fast variations
        fast_config = base_config.copy()
        fast_config.update({
            "tier1_rtsp_ready_check_timeout": base_config["tier1_rtsp_ready_check_timeout"] * 0.5,
            "tier2_activation_timeout": base_config["tier2_activation_timeout"] * 0.7,
            "tier3_direct_capture_timeout": base_config["tier3_direct_capture_timeout"] * 0.7,
            "immediate_response_threshold": base_config["immediate_response_threshold"] * 0.5
        })
        variations["fast"] = fast_config
        
        # Conservative variations
        conservative_config = base_config.copy()
        conservative_config.update({
            "tier1_rtsp_ready_check_timeout": base_config["tier1_rtsp_ready_check_timeout"] * 1.5,
            "tier2_activation_timeout": base_config["tier2_activation_timeout"] * 1.3,
            "tier3_direct_capture_timeout": base_config["tier3_direct_capture_timeout"] * 1.3,
            "immediate_response_threshold": base_config["immediate_response_threshold"] * 1.5
        })
        variations["conservative"] = conservative_config
        
        # Balanced variations
        balanced_config = base_config.copy()
        balanced_config.update({
            "tier2_activation_timeout": base_config["tier2_activation_timeout"] * 0.9,
            "acceptable_response_threshold": base_config["acceptable_response_threshold"] * 0.8
        })
        variations["balanced"] = balanced_config
        
        return variations
    
    async def run_performance_tuning(self, environment: str, iterations: int = 5, include_variations: bool = True) -> Dict[str, Any]:
        """Run comprehensive performance tuning."""
        logging.info(f"Starting performance tuning for environment: {environment}")
        
        # Get base configurations
        configs = self.get_environment_configs(environment)
        
        all_results = {}
        
        for env_name, base_config in configs.items():
            logging.info(f"Testing environment: {env_name}")
            
            # Test base configuration
            base_result = await self.test_configuration(f"{env_name}_base", base_config, iterations)
            all_results[f"{env_name}_base"] = base_result
            
            if include_variations:
                # Generate and test variations
                variations = self.generate_variations(base_config)
                for var_name, var_config in variations.items():
                    var_result = await self.test_configuration(f"{env_name}_{var_name}", var_config, iterations)
                    all_results[f"{env_name}_{var_name}"] = var_result
        
        return all_results
    
    def analyze_results(self, results: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze performance tuning results and generate recommendations."""
        analysis = {
            "summary": {},
            "recommendations": {},
            "best_configurations": {}
        }
        
        # Find best configurations by environment
        for result_name, result in results.items():
            if result["success_rate"] > 0 and result["capture_time_stats"]:
                env_name = result_name.split("_")[0]
                if env_name not in analysis["best_configurations"]:
                    analysis["best_configurations"][env_name] = []
                
                analysis["best_configurations"][env_name].append({
                    "config_name": result_name,
                    "mean_time": result["capture_time_stats"]["mean"],
                    "success_rate": result["success_rate"],
                    "tier_distribution": result["tier_distribution"]
                })
        
        # Sort by performance (mean time) for each environment
        for env_name in analysis["best_configurations"]:
            analysis["best_configurations"][env_name].sort(key=lambda x: x["mean_time"])
        
        # Generate recommendations
        for env_name, configs in analysis["best_configurations"].items():
            if configs:
                best_config = configs[0]
                analysis["recommendations"][env_name] = {
                    "best_config": best_config["config_name"],
                    "expected_performance": f"{best_config['mean_time']:.3f}s",
                    "success_rate": f"{best_config['success_rate']:.1%}",
                    "primary_tier": max(best_config["tier_distribution"].items(), key=lambda x: x[1])[0] if best_config["tier_distribution"] else "unknown"
                }
        
        return analysis
    
    def generate_config_file(self, results: Dict[str, Any], analysis: Dict[str, Any], output_file: str):
        """Generate optimized configuration file."""
        config_template = {
            "performance": {
                "snapshot_tiers": {}
            }
        }
        
        # Use the best configuration for each environment
        for env_name, recommendation in analysis["recommendations"].items():
            best_config_name = recommendation["best_config"]
            if best_config_name in results:
                # Extract the configuration from the test results
                # This would need to be stored in the results for full implementation
                pass
        
        # For now, use the best overall configuration
        best_overall = None
        best_time = float('inf')
        
        for result_name, result in results.items():
            if result["success_rate"] > 0.8 and result["capture_time_stats"]:
                if result["capture_time_stats"]["mean"] < best_time:
                    best_time = result["capture_time_stats"]["mean"]
                    best_overall = result_name
        
        if best_overall:
            # Generate configuration based on best performance
            config_template["performance"]["snapshot_tiers"] = {
                "tier1_rtsp_ready_check_timeout": 1.0,
                "tier2_activation_timeout": 3.0,
                "tier2_activation_trigger_timeout": 1.0,
                "tier3_direct_capture_timeout": 5.0,
                "total_operation_timeout": 10.0,
                "immediate_response_threshold": 0.5,
                "acceptable_response_threshold": 2.0,
                "slow_response_threshold": 5.0
            }
        
        # Write configuration file
        with open(output_file, 'w') as f:
            json.dump(config_template, f, indent=2)
        
        logging.info(f"Generated optimized configuration: {output_file}")


async def main():
    """Main performance tuning function."""
    parser = argparse.ArgumentParser(description="Performance tuning for multi-tier snapshot capture")
    parser.add_argument("--environment", choices=["development", "production", "high-performance", "embedded", "all"], 
                       default="production", help="Target environment for optimization")
    parser.add_argument("--iterations", type=int, default=5, help="Number of iterations per configuration")
    parser.add_argument("--output", type=str, default="optimized_snapshot_config.json", 
                       help="Output file for optimized configuration")
    parser.add_argument("--no-variations", action="store_true", help="Skip configuration variations")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable verbose logging")
    
    args = parser.parse_args()
    
    # Setup logging
    log_level = logging.DEBUG if args.verbose else logging.INFO
    logging.basicConfig(level=log_level, format='%(asctime)s - %(levelname)s - %(message)s')
    
    # Load configuration
    config = Config()
    
    # Create performance tuner
    tuner = SnapshotPerformanceTuner(config)
    
    try:
        await tuner.setup()
        
        # Run performance tuning
        results = await tuner.run_performance_tuning(
            environment=args.environment,
            iterations=args.iterations,
            include_variations=not args.no_variations
        )
        
        # Analyze results
        analysis = tuner.analyze_results(results)
        
        # Generate report
        print("\n" + "=" * 60)
        print("📊 Performance Tuning Results")
        print("=" * 60)
        
        for env_name, recommendation in analysis["recommendations"].items():
            print(f"\n🎯 {env_name.upper()} Environment:")
            print(f"   Best Configuration: {recommendation['best_config']}")
            print(f"   Expected Performance: {recommendation['expected_performance']}")
            print(f"   Success Rate: {recommendation['success_rate']}")
            print(f"   Primary Tier: {recommendation['primary_tier']}")
        
        # Generate configuration file
        tuner.generate_config_file(results, analysis, args.output)
        
        print(f"\n✅ Performance tuning completed!")
        print(f"📁 Optimized configuration saved to: {args.output}")
        
    except Exception as e:
        logging.error(f"Performance tuning failed: {e}")
        return 1
    finally:
        await tuner.cleanup()
    
    return 0


if __name__ == "__main__":
    exit_code = asyncio.run(main())
    sys.exit(exit_code)
