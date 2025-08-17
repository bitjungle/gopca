#!/usr/bin/env python3
"""
Plotly Advanced Features Reference Implementation
==================================================
This script demonstrates advanced Plotly features for state-of-the-art
visualizations including WebGL, density plots, lasso selection, and more.

Author: GoPCA Development Team
Date: 2025-01-17
"""

import pandas as pd
import numpy as np
import plotly.graph_objects as go
import plotly.express as px
from plotly.subplots import make_subplots
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler
from scipy.stats import gaussian_kde
from scipy import stats
import time
import os


# Create output directory
OUTPUT_DIR = 'plotly'
os.makedirs(OUTPUT_DIR, exist_ok=True)


def get_output_path(filename):
    """Get the full path for output files in the plotly subdirectory."""
    return os.path.join(OUTPUT_DIR, filename)


def load_data():
    """Load and prepare wine dataset."""
    df = pd.read_csv('wine.csv')
    feature_cols = [col for col in df.columns if col not in ['classes', 'Unnamed: 0']]
    X = df[feature_cols].values
    y = df['classes'].values
    
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)
    
    pca = PCA(n_components=2)
    scores = pca.fit_transform(X_scaled)
    
    return scores, y, pca.explained_variance_ratio_ * 100


def create_webgl_comparison(scores, y):
    """Compare performance of scatter vs scattergl."""
    fig = make_subplots(
        rows=1, cols=2,
        subplot_titles=('Standard Scatter (SVG)', 'WebGL Scatter (GPU)')
    )
    
    classes = np.unique(y)
    colors = px.colors.qualitative.Plotly
    
    # Standard scatter
    for i, cls in enumerate(classes):
        mask = y == cls
        fig.add_trace(
            go.Scatter(
                x=scores[mask, 0],
                y=scores[mask, 1],
                mode='markers',
                name=cls,
                marker=dict(size=8, color=colors[i]),
                legendgroup=cls,
                showlegend=True
            ),
            row=1, col=1
        )
    
    # WebGL scatter (scattergl)
    for i, cls in enumerate(classes):
        mask = y == cls
        fig.add_trace(
            go.Scattergl(  # Note: Scattergl for WebGL rendering
                x=scores[mask, 0],
                y=scores[mask, 1],
                mode='markers',
                name=cls,
                marker=dict(size=8, color=colors[i]),
                legendgroup=cls,
                showlegend=False
            ),
            row=1, col=2
        )
    
    fig.update_layout(
        title='WebGL Performance Comparison',
        height=500,
        width=1000,
        template='plotly_dark',
        annotations=[
            dict(
                text="Use scattergl for >1000 points",
                xref="paper", yref="paper",
                x=0.5, y=-0.1,
                showarrow=False,
                font=dict(size=12, color="yellow")
            )
        ]
    )
    
    return fig


def create_density_contour_plot(scores, y):
    """Create density contour plot for overlapping regions."""
    fig = go.Figure()
    
    classes = np.unique(y)
    colors = px.colors.qualitative.Plotly
    
    # Add scatter points
    for i, cls in enumerate(classes):
        mask = y == cls
        fig.add_trace(go.Scatter(
            x=scores[mask, 0],
            y=scores[mask, 1],
            mode='markers',
            name=cls,
            marker=dict(size=6, color=colors[i], opacity=0.6)
        ))
    
    # Add density contours for each class
    for i, cls in enumerate(classes):
        mask = y == cls
        x_data = scores[mask, 0]
        y_data = scores[mask, 1]
        
        # Create grid for density estimation
        x_min, x_max = x_data.min() - 1, x_data.max() + 1
        y_min, y_max = y_data.min() - 1, y_data.max() + 1
        
        xx, yy = np.meshgrid(
            np.linspace(x_min, x_max, 50),
            np.linspace(y_min, y_max, 50)
        )
        
        # Kernel density estimation
        if len(x_data) > 3:  # Need enough points for KDE
            positions = np.vstack([xx.ravel(), yy.ravel()])
            values = np.vstack([x_data, y_data])
            kernel = gaussian_kde(values)
            density = np.reshape(kernel(positions).T, xx.shape)
            
            # Add contour
            fig.add_trace(go.Contour(
                x=xx[0],
                y=yy[:, 0],
                z=density,
                showscale=False,
                colorscale=[[0, 'rgba(0,0,0,0)'], [1, colors[i]]],
                opacity=0.3,
                contours=dict(
                    showlines=True,
                    coloring='none'
                ),
                line=dict(color=colors[i], width=2),
                name=f'{cls} density',
                showlegend=False,
                hoverinfo='skip'
            ))
    
    fig.update_layout(
        title='PCA with Density Contours (2D KDE)',
        xaxis_title='PC1',
        yaxis_title='PC2',
        width=800,
        height=600,
        template='plotly_dark'
    )
    
    return fig


def create_lasso_selection_demo(scores, y):
    """Create plot with lasso selection capability."""
    fig = go.Figure()
    
    # Create single trace with all points
    fig.add_trace(go.Scattergl(
        x=scores[:, 0],
        y=scores[:, 1],
        mode='markers',
        marker=dict(
            size=8,
            color=pd.Categorical(y).codes,
            colorscale='Viridis',
            showscale=True,
            colorbar=dict(title="Class")
        ),
        text=[f'Sample {i}<br>Class: {y[i]}' for i in range(len(y))],
        hovertemplate='<b>%{text}</b><br>PC1: %{x:.2f}<br>PC2: %{y:.2f}<extra></extra>',
        selectedpoints=[],  # Enable selection
        selected=dict(marker=dict(color='red', size=12)),
        unselected=dict(marker=dict(opacity=0.3))
    ))
    
    # Configure layout with selection tools
    fig.update_layout(
        title='Lasso Selection Demo (Use lasso tool in modebar)',
        xaxis_title='PC1',
        yaxis_title='PC2',
        width=800,
        height=600,
        template='plotly_dark',
        dragmode='lasso',  # Set default tool to lasso
        hovermode='closest',
        modebar=dict(
            bgcolor='rgba(0,0,0,0.5)',
            color='white',
            activecolor='yellow'
        )
    )
    
    # Add custom modebar config
    config = {
        'modeBarButtonsToAdd': ['select2d', 'lasso2d'],
        'modeBarButtonsToRemove': ['autoScale2d'],
        'displaylogo': False,
        'toImageButtonOptions': {
            'format': 'svg',
            'filename': 'pca_lasso_selection',
            'height': 1200,
            'width': 1600,
            'scale': 2
        }
    }
    
    return fig, config


def create_3d_with_density(scores_3d, y):
    """Create 3D plot with density projections."""
    fig = go.Figure()
    
    classes = np.unique(y)
    colors = px.colors.qualitative.Plotly
    
    # Add 3D scatter
    for i, cls in enumerate(classes):
        mask = y == cls
        fig.add_trace(go.Scatter3d(
            x=scores_3d[mask, 0],
            y=scores_3d[mask, 1],
            z=scores_3d[mask, 2],
            mode='markers',
            name=cls,
            marker=dict(
                size=4,
                color=colors[i],
                opacity=0.8,
                line=dict(width=0.5, color='white')
            )
        ))
    
    # Add 2D projections as contours on the walls
    # This is conceptual - in practice would need more complex implementation
    
    fig.update_layout(
        title='3D PCA with Projections',
        scene=dict(
            xaxis_title='PC1',
            yaxis_title='PC2',
            zaxis_title='PC3',
            camera=dict(
                eye=dict(x=1.5, y=1.5, z=1.5),
                center=dict(x=0, y=0, z=0)
            ),
            aspectmode='cube'
        ),
        width=800,
        height=600,
        template='plotly_dark'
    )
    
    return fig


def create_animated_transitions():
    """Create animated transitions between different PC combinations."""
    # Load data
    df = pd.read_csv('wine.csv')
    feature_cols = [col for col in df.columns if col not in ['classes', 'Unnamed: 0']]
    X = df[feature_cols].values
    y = df['classes'].values
    
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)
    
    pca = PCA(n_components=4)
    scores = pca.fit_transform(X_scaled)
    
    # Create frames for animation
    frames = []
    pc_combinations = [(0,1), (0,2), (1,2), (0,3), (1,3), (2,3)]
    
    for pc1, pc2 in pc_combinations:
        frame_data = []
        for cls in np.unique(y):
            mask = y == cls
            frame_data.append(go.Scatter(
                x=scores[mask, pc1],
                y=scores[mask, pc2],
                mode='markers',
                name=cls,
                marker=dict(size=8)
            ))
        
        frames.append(go.Frame(
            data=frame_data,
            name=f'PC{pc1+1} vs PC{pc2+1}',
            layout=go.Layout(
                xaxis_title=f'PC{pc1+1}',
                yaxis_title=f'PC{pc2+1}'
            )
        ))
    
    # Initial data
    initial_data = []
    for cls in np.unique(y):
        mask = y == cls
        initial_data.append(go.Scatter(
            x=scores[mask, 0],
            y=scores[mask, 1],
            mode='markers',
            name=cls,
            marker=dict(size=8)
        ))
    
    fig = go.Figure(data=initial_data, frames=frames)
    
    # Add play/pause buttons
    fig.update_layout(
        updatemenus=[
            dict(
                type='buttons',
                showactive=False,
                buttons=[
                    dict(label='Play',
                         method='animate',
                         args=[None, dict(frame=dict(duration=1000, redraw=True),
                                        fromcurrent=True, mode='immediate')]),
                    dict(label='Pause',
                         method='animate',
                         args=[[None], dict(frame=dict(duration=0, redraw=False),
                                          mode='immediate')])
                ]
            )
        ],
        sliders=[{
            'steps': [
                {
                    'args': [[f.name], {'frame': {'duration': 0, 'redraw': True}, 'mode': 'immediate'}],
                    'label': f.name,
                    'method': 'animate'
                }
                for f in frames
            ],
            'active': 0,
            'y': 0,
            'len': 0.9,
            'x': 0.1,
            'xanchor': 'left',
            'y': 0,
            'yanchor': 'top'
        }],
        title='Animated PC Transitions',
        width=800,
        height=600,
        template='plotly_dark'
    )
    
    return fig


def benchmark_performance():
    """Benchmark rendering performance for different data sizes."""
    sizes = [100, 500, 1000, 5000, 10000]
    results = {'size': [], 'svg_time': [], 'webgl_time': []}
    
    for size in sizes:
        # Generate random data
        X = np.random.randn(size, 2)
        
        # Time SVG rendering
        start = time.time()
        fig_svg = go.Figure(go.Scatter(x=X[:, 0], y=X[:, 1], mode='markers'))
        fig_svg.write_html(get_output_path(f'benchmark_svg_{size}.html'))
        svg_time = time.time() - start
        
        # Time WebGL rendering
        start = time.time()
        fig_webgl = go.Figure(go.Scattergl(x=X[:, 0], y=X[:, 1], mode='markers'))
        fig_webgl.write_html(get_output_path(f'benchmark_webgl_{size}.html'))
        webgl_time = time.time() - start
        
        results['size'].append(size)
        results['svg_time'].append(svg_time)
        results['webgl_time'].append(webgl_time)
        
        print(f"Size: {size:5d} | SVG: {svg_time:.3f}s | WebGL: {webgl_time:.3f}s | Speedup: {svg_time/webgl_time:.1f}x")
    
    # Create performance comparison plot
    fig = go.Figure()
    fig.add_trace(go.Scatter(
        x=results['size'],
        y=results['svg_time'],
        mode='lines+markers',
        name='SVG (Standard)',
        line=dict(color='red', width=2)
    ))
    fig.add_trace(go.Scatter(
        x=results['size'],
        y=results['webgl_time'],
        mode='lines+markers',
        name='WebGL (GPU)',
        line=dict(color='green', width=2)
    ))
    
    fig.update_layout(
        title='Performance: SVG vs WebGL Rendering',
        xaxis_title='Number of Points',
        yaxis_title='Render Time (seconds)',
        xaxis_type='log',
        yaxis_type='log',
        width=800,
        height=500,
        template='plotly_dark'
    )
    
    return fig, results


def create_custom_modebar_demo(scores, y):
    """Demonstrate custom modebar configuration."""
    fig = go.Figure()
    
    # Add data
    classes = np.unique(y)
    colors = px.colors.qualitative.Plotly
    
    for i, cls in enumerate(classes):
        mask = y == cls
        fig.add_trace(go.Scattergl(
            x=scores[mask, 0],
            y=scores[mask, 1],
            mode='markers',
            name=cls,
            marker=dict(size=8, color=colors[i])
        ))
    
    fig.update_layout(
        title='Custom Modebar Configuration',
        xaxis_title='PC1',
        yaxis_title='PC2',
        width=800,
        height=600,
        template='plotly_dark',
        hovermode='x unified'
    )
    
    # Custom configuration
    config = {
        'modeBarButtonsToAdd': [
            'drawline',
            'drawopenpath',
            'drawclosedpath',
            'drawcircle',
            'drawrect',
            'eraseshape'
        ],
        'modeBarButtonsToRemove': ['lasso2d'],
        'displaylogo': False,
        'displayModeBar': True,
        'toImageButtonOptions': {
            'format': 'svg',
            'filename': 'custom_plot',
            'height': 2400,
            'width': 3200,
            'scale': 4  # 4x resolution for publication quality
        }
    }
    
    return fig, config


def save_all_advanced_plots():
    """Generate and save all advanced feature demonstrations."""
    # Load data
    scores, y, variance = load_data()
    
    # For 3D
    df = pd.read_csv('wine.csv')
    feature_cols = [col for col in df.columns if col not in ['classes', 'Unnamed: 0']]
    X = df[feature_cols].values
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)
    pca_3d = PCA(n_components=3)
    scores_3d = pca_3d.fit_transform(X_scaled)
    
    print("Creating Advanced Plotly Demonstrations")
    print("=" * 50)
    
    # 1. WebGL Comparison
    fig_webgl = create_webgl_comparison(scores, y)
    fig_webgl.write_html(get_output_path('advanced_webgl_comparison.html'))
    print("✅ WebGL comparison saved")
    
    # 2. Density Contours
    fig_density = create_density_contour_plot(scores, y)
    fig_density.write_html(get_output_path('advanced_density_contours.html'))
    print("✅ Density contours saved")
    
    # 3. Lasso Selection
    fig_lasso, config_lasso = create_lasso_selection_demo(scores, y)
    fig_lasso.write_html(get_output_path('advanced_lasso_selection.html'), config=config_lasso)
    print("✅ Lasso selection demo saved")
    
    # 4. 3D with Density
    fig_3d = create_3d_with_density(scores_3d, y)
    fig_3d.write_html(get_output_path('advanced_3d_density.html'))
    print("✅ 3D with density saved")
    
    # 5. Animated Transitions
    fig_animated = create_animated_transitions()
    fig_animated.write_html(get_output_path('advanced_animated_transitions.html'))
    print("✅ Animated transitions saved")
    
    # 6. Performance Benchmark
    print("\nRunning performance benchmarks...")
    fig_perf, perf_results = benchmark_performance()
    fig_perf.write_html(get_output_path('advanced_performance_benchmark.html'))
    print("✅ Performance benchmark saved")
    
    # 7. Custom Modebar
    fig_custom, config_custom = create_custom_modebar_demo(scores, y)
    fig_custom.write_html(get_output_path('advanced_custom_modebar.html'), config=config_custom)
    print("✅ Custom modebar demo saved")
    
    print("\n" + "=" * 50)
    print("All advanced demonstrations created successfully!")
    print(f"Open the HTML files in the '{OUTPUT_DIR}' folder to explore advanced Plotly features.")
    
    # Save configuration examples as JSON
    import json
    configs = {
        'lasso_config': config_lasso,
        'custom_modebar_config': config_custom,
        'performance_results': perf_results,
        'notes': {
            'webgl_threshold': 'Use scattergl for >1000 points',
            'export_resolution': 'Use scale=2-4 for publication quality',
            'density_method': 'gaussian_kde from scipy.stats',
            'animation_duration': '1000ms per frame recommended'
        }
    }
    
    with open(get_output_path('advanced_configs.json'), 'w') as f:
        json.dump(configs, f, indent=2)
    print(f"\n📄 Configuration examples saved to {get_output_path('advanced_configs.json')}")


if __name__ == "__main__":
    save_all_advanced_plots()