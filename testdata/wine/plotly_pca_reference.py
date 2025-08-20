#!/usr/bin/env python3
"""
Plotly.js Reference Implementation for PCA Visualizations
==========================================================
This script creates reference PCA visualizations using Plotly Python
to guide the JavaScript implementation for GoPCA.

Author: GoPCA Development Team
Date: 2025-01-17
Dataset: Wine dataset (wine.csv)

References:
- Johnson, R. A., & Wichern, D. W. (2007). Applied Multivariate Statistical Analysis (6th ed.)
- Gabriel, K. R. (1971). The biplot graphic display. Biometrika, 58(3), 453-467
"""

import pandas as pd
import numpy as np
import plotly.graph_objects as go
import plotly.express as px
from plotly.subplots import make_subplots
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler
from scipy import stats
import json
import os


# Create output directory
OUTPUT_DIR = 'plotly'
os.makedirs(OUTPUT_DIR, exist_ok=True)


def get_output_path(filename):
    """Get the full path for output files in the plotly subdirectory."""
    return os.path.join(OUTPUT_DIR, filename)


def load_and_prepare_data():
    """Load wine dataset and prepare for PCA."""
    df = pd.read_csv('wine.csv')
    
    # Separate features and classes
    feature_cols = [col for col in df.columns if col not in ['classes', 'Unnamed: 0']]
    X = df[feature_cols].values
    y = df['classes'].values
    feature_names = feature_cols
    
    # Standardize features
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)
    
    return X_scaled, y, feature_names, df


def perform_pca(X_scaled, n_components=None):
    """Perform PCA and return results."""
    pca = PCA(n_components=n_components)
    scores = pca.fit_transform(X_scaled)
    
    return {
        'scores': scores,
        'loadings': pca.components_.T,
        'explained_variance': pca.explained_variance_,
        'explained_variance_ratio': pca.explained_variance_ratio_ * 100,
        'cumulative_variance': np.cumsum(pca.explained_variance_ratio_) * 100,
        'n_components': pca.n_components_,
        'mean': pca.mean_,
        'components': pca.components_
    }


def calculate_confidence_ellipse(x, y, confidence=0.95):
    """
    Calculate confidence ellipse parameters.
    Reference: Johnson & Wichern (2007), Ch. 4
    """
    mean_x, mean_y = np.mean(x), np.mean(y)
    cov_matrix = np.cov(x, y)
    
    # Chi-square value for confidence level
    chi2_val = stats.chi2.ppf(confidence, df=2)
    
    # Eigenvalues and eigenvectors for ellipse orientation
    eigenvalues, eigenvectors = np.linalg.eig(cov_matrix)
    
    # Ellipse parameters
    angle = np.degrees(np.arctan2(eigenvectors[1, 0], eigenvectors[0, 0]))
    width = 2 * np.sqrt(chi2_val * eigenvalues[0])
    height = 2 * np.sqrt(chi2_val * eigenvalues[1])
    
    # Generate ellipse points
    theta = np.linspace(0, 2*np.pi, 100)
    ellipse_x = width/2 * np.cos(theta)
    ellipse_y = height/2 * np.sin(theta)
    
    # Rotate ellipse
    rotation_matrix = np.array([
        [np.cos(np.radians(angle)), -np.sin(np.radians(angle))],
        [np.sin(np.radians(angle)), np.cos(np.radians(angle))]
    ])
    
    ellipse_points = np.dot(rotation_matrix, np.array([ellipse_x, ellipse_y]))
    ellipse_x = ellipse_points[0, :] + mean_x
    ellipse_y = ellipse_points[1, :] + mean_y
    
    return ellipse_x, ellipse_y


def calculate_smart_labels(scores, max_labels=10):
    """
    Select points furthest from origin for labeling.
    This preserves the beloved smart label selection feature.
    """
    distances = np.sqrt(scores[:, 0]**2 + scores[:, 1]**2)
    top_indices = np.argsort(distances)[-max_labels:]
    return top_indices


def create_scores_plot(pca_results, y, pc1=0, pc2=1, show_ellipses=True, max_labels=10):
    """Create interactive scores plot with smart labels and confidence ellipses."""
    fig = go.Figure()
    
    scores = pca_results['scores']
    classes = np.unique(y)
    colors = px.colors.qualitative.Plotly
    
    # Calculate smart labels
    smart_label_indices = calculate_smart_labels(scores[:, [pc1, pc2]], max_labels)
    
    for i, cls in enumerate(classes):
        mask = y == cls
        class_scores = scores[mask]
        
        # Create text array - only show labels for smart selection
        text = [''] * len(class_scores)
        indices = np.where(mask)[0]
        for j, idx in enumerate(indices):
            if idx in smart_label_indices:
                text[j] = f'Sample {idx}'
        
        # Add scatter trace
        fig.add_trace(go.Scatter(
            x=class_scores[:, pc1],
            y=class_scores[:, pc2],
            mode='markers+text',
            name=cls,
            text=text,
            textposition='top center',
            marker=dict(size=8, color=colors[i % len(colors)]),
            hovertemplate=f'<b>%{{text}}</b><br>PC{pc1+1}: %{{x:.2f}}<br>PC{pc2+1}: %{{y:.2f}}<extra></extra>'
        ))
        
        # Add confidence ellipse
        if show_ellipses and len(class_scores) > 2:
            ellipse_x, ellipse_y = calculate_confidence_ellipse(
                class_scores[:, pc1], 
                class_scores[:, pc2], 
                confidence=0.95
            )
            fig.add_trace(go.Scatter(
                x=ellipse_x,
                y=ellipse_y,
                mode='lines',
                line=dict(color=colors[i % len(colors)], dash='dash', width=2),
                showlegend=False,
                hoverinfo='skip'
            ))
    
    # Update layout
    fig.update_layout(
        title='PCA Scores Plot with Smart Labels and Confidence Ellipses',
        xaxis_title=f"PC{pc1+1} ({pca_results['explained_variance_ratio'][pc1]:.1f}%)",
        yaxis_title=f"PC{pc2+1} ({pca_results['explained_variance_ratio'][pc2]:.1f}%)",
        hovermode='closest',
        width=800,
        height=600,
        template='plotly_dark'
    )
    
    # Add reference lines at origin
    fig.add_hline(y=0, line_dash="solid", line_color="gray", opacity=0.5)
    fig.add_vline(x=0, line_dash="solid", line_color="gray", opacity=0.5)
    
    return fig


def create_scree_plot(pca_results):
    """Create scree plot with dual y-axis."""
    fig = make_subplots(specs=[[{"secondary_y": True}]])
    
    n_components = pca_results['n_components']
    x = list(range(1, n_components + 1))
    
    # Bar chart for explained variance
    fig.add_trace(
        go.Bar(
            x=x,
            y=pca_results['explained_variance_ratio'],
            name='Explained Variance',
            marker_color=px.colors.qualitative.Plotly[:n_components]
        ),
        secondary_y=False
    )
    
    # Line chart for cumulative variance
    fig.add_trace(
        go.Scatter(
            x=x,
            y=pca_results['cumulative_variance'],
            mode='lines+markers',
            name='Cumulative Variance',
            line=dict(color='green', width=2),
            marker=dict(size=8)
        ),
        secondary_y=True
    )
    
    # Add 80% threshold line
    fig.add_hline(
        y=80,
        line_dash="dash",
        line_color="red",
        annotation_text="80%",
        secondary_y=True
    )
    
    # Update layout
    fig.update_xaxes(title_text="Principal Component", tickmode='linear', dtick=1)
    fig.update_yaxes(title_text="Explained Variance (%)", secondary_y=False)
    fig.update_yaxes(title_text="Cumulative Variance (%)", secondary_y=True, range=[0, 105])
    
    fig.update_layout(
        title='Scree Plot with Cumulative Variance',
        hovermode='x unified',
        width=800,
        height=500,
        template='plotly_dark'
    )
    
    return fig


def create_loadings_plot(pca_results, feature_names, pc=0, plot_type='bar'):
    """Create loadings plot (bar or line chart)."""
    loadings = pca_results['loadings'][:, pc]
    
    if plot_type == 'bar':
        colors = ['red' if l < 0 else 'blue' for l in loadings]
        fig = go.Figure(go.Bar(
            x=feature_names,
            y=loadings,
            marker_color=colors,
            hovertemplate='<b>%{x}</b><br>Loading: %{y:.3f}<extra></extra>'
        ))
        chart_type = 'Bar Chart'
    else:
        fig = go.Figure(go.Scatter(
            x=list(range(len(feature_names))),
            y=loadings,
            mode='lines+markers',
            line=dict(color='blue', width=2),
            marker=dict(size=8),
            text=feature_names,
            hovertemplate='<b>%{text}</b><br>Loading: %{y:.3f}<extra></extra>'
        ))
        chart_type = 'Line Chart'
    
    fig.add_hline(y=0, line_dash="solid", line_color="gray", opacity=0.5)
    
    fig.update_layout(
        title=f'PC{pc+1} Loadings ({chart_type}) - {pca_results["explained_variance_ratio"][pc]:.1f}% variance',
        xaxis_title='Variable' if plot_type == 'bar' else 'Variable Index',
        yaxis_title='Loading Value',
        width=800,
        height=500,
        template='plotly_dark'
    )
    
    if plot_type == 'bar':
        fig.update_xaxes(tickangle=-45)
    
    return fig


def create_biplot(pca_results, y, feature_names, pc1=0, pc2=1, scale=0.5):
    """
    Create biplot with properly scaled vectors.
    Reference: Gabriel (1971) - scale parameter controls emphasis
    scale=0: row-metric preserving
    scale=1: column-metric preserving  
    scale=0.5: symmetric (default)
    """
    fig = go.Figure()
    
    scores = pca_results['scores']
    loadings = pca_results['loadings']
    
    # Scale factors (Gabriel 1971, Equation 2)
    n_samples = scores.shape[0]
    n_features = loadings.shape[0]
    alpha = np.power(n_samples, scale)
    beta = np.power(n_features, 1 - scale)
    
    # Scaled coordinates
    scaled_scores = scores * beta
    scaled_loadings = loadings * alpha
    
    # Add scores
    classes = np.unique(y)
    colors = px.colors.qualitative.Plotly
    
    for i, cls in enumerate(classes):
        mask = y == cls
        fig.add_trace(go.Scatter(
            x=scaled_scores[mask, pc1],
            y=scaled_scores[mask, pc2],
            mode='markers',
            name=cls,
            marker=dict(size=6, color=colors[i % len(colors)], opacity=0.7)
        ))
    
    # Add loading vectors
    max_score = np.max(np.abs(scaled_scores[:, [pc1, pc2]]))
    scale_factor = max_score * 0.8 / np.max(np.abs(scaled_loadings[:, [pc1, pc2]]))
    
    for i, feature in enumerate(feature_names):
        fig.add_annotation(
            x=scaled_loadings[i, pc1] * scale_factor,
            y=scaled_loadings[i, pc2] * scale_factor,
            ax=0, ay=0,
            xref='x', yref='y',
            axref='x', ayref='y',
            showarrow=True,
            arrowhead=2,
            arrowsize=1,
            arrowwidth=2,
            arrowcolor='red',
            text=feature,
            font=dict(size=10),
            bgcolor='rgba(255,255,255,0.7)',
            bordercolor='red',
            borderwidth=1
        )
    
    fig.update_layout(
        title=f'Biplot (Gabriel scaling, α={scale})',
        xaxis_title=f"PC{pc1+1} ({pca_results['explained_variance_ratio'][pc1]:.1f}%)",
        yaxis_title=f"PC{pc2+1} ({pca_results['explained_variance_ratio'][pc2]:.1f}%)",
        width=800,
        height=600,
        template='plotly_dark'
    )
    
    # Add reference lines
    fig.add_hline(y=0, line_dash="solid", line_color="gray", opacity=0.5)
    fig.add_vline(x=0, line_dash="solid", line_color="gray", opacity=0.5)
    
    return fig


def create_3d_scores_plot(pca_results, y, pc1=0, pc2=1, pc3=2):
    """Create interactive 3D scores plot."""
    scores = pca_results['scores']
    
    fig = go.Figure()
    
    classes = np.unique(y)
    colors = px.colors.qualitative.Plotly
    
    for i, cls in enumerate(classes):
        mask = y == cls
        fig.add_trace(go.Scatter3d(
            x=scores[mask, pc1],
            y=scores[mask, pc2],
            z=scores[mask, pc3],
            mode='markers',
            name=cls,
            marker=dict(
                size=5,
                color=colors[i % len(colors)],
                opacity=0.8
            ),
            text=[f'Sample {j}' for j in range(sum(mask))],
            hovertemplate='<b>%{text}</b><br>' +
                         f'PC{pc1+1}: %{{x:.2f}}<br>' +
                         f'PC{pc2+1}: %{{y:.2f}}<br>' +
                         f'PC{pc3+1}: %{{z:.2f}}<extra></extra>'
        ))
    
    fig.update_layout(
        title='3D PCA Scores Plot',
        scene=dict(
            xaxis_title=f"PC{pc1+1} ({pca_results['explained_variance_ratio'][pc1]:.1f}%)",
            yaxis_title=f"PC{pc2+1} ({pca_results['explained_variance_ratio'][pc2]:.1f}%)",
            zaxis_title=f"PC{pc3+1} ({pca_results['explained_variance_ratio'][pc3]:.1f}%)",
            camera=dict(
                eye=dict(x=1.5, y=1.5, z=1.5)
            )
        ),
        width=800,
        height=600,
        template='plotly_dark'
    )
    
    return fig


def create_circle_of_correlations(pca_results, feature_names, pc1=0, pc2=1):
    """Create circle of correlations plot."""
    loadings = pca_results['loadings']
    
    fig = go.Figure()
    
    # Add unit circle
    theta = np.linspace(0, 2*np.pi, 100)
    fig.add_trace(go.Scatter(
        x=np.cos(theta),
        y=np.sin(theta),
        mode='lines',
        line=dict(color='gray', dash='dot'),
        showlegend=False,
        hoverinfo='skip'
    ))
    
    # Add loading vectors
    for i, feature in enumerate(feature_names):
        fig.add_annotation(
            x=loadings[i, pc1],
            y=loadings[i, pc2],
            ax=0, ay=0,
            xref='x', yref='y',
            axref='x', ayref='y',
            showarrow=True,
            arrowhead=2,
            arrowsize=1,
            arrowwidth=2,
            arrowcolor='red',
            text=feature,
            font=dict(size=9),
            bgcolor='rgba(255,255,255,0.8)',
            bordercolor='red',
            borderwidth=1
        )
    
    fig.update_layout(
        title='Circle of Correlations',
        xaxis_title=f"PC{pc1+1} ({pca_results['explained_variance_ratio'][pc1]:.1f}%)",
        yaxis_title=f"PC{pc2+1} ({pca_results['explained_variance_ratio'][pc2]:.1f}%)",
        xaxis=dict(range=[-1.2, 1.2], zeroline=True, zerolinecolor='gray'),
        yaxis=dict(range=[-1.2, 1.2], zeroline=True, zerolinecolor='gray'),
        width=600,
        height=600,
        template='plotly_dark'
    )
    
    fig.update_xaxes(constrain="domain")
    fig.update_yaxes(scaleanchor="x", scaleratio=1, constrain="domain")
    
    return fig


def save_all_plots():
    """Generate and save all PCA visualizations."""
    # Load data and perform PCA
    X_scaled, y, feature_names, df = load_and_prepare_data()
    pca_results = perform_pca(X_scaled)
    
    # Print PCA summary
    print("PCA Results Summary")
    print("=" * 50)
    print(f"Number of samples: {X_scaled.shape[0]}")
    print(f"Number of features: {X_scaled.shape[1]}")
    print(f"Number of components: {pca_results['n_components']}")
    print(f"\nExplained Variance Ratio:")
    for i, var in enumerate(pca_results['explained_variance_ratio'][:5]):
        print(f"  PC{i+1}: {var:.1f}%")
    print(f"\nCumulative Variance (first 5 PCs): {pca_results['cumulative_variance'][4]:.1f}%")
    
    # Create all plots
    plots = {
        'scores_plot': create_scores_plot(pca_results, y),
        'scree_plot': create_scree_plot(pca_results),
        'loadings_bar': create_loadings_plot(pca_results, feature_names, pc=0, plot_type='bar'),
        'loadings_line': create_loadings_plot(pca_results, feature_names, pc=0, plot_type='line'),
        'biplot': create_biplot(pca_results, y, feature_names),
        '3d_scores': create_3d_scores_plot(pca_results, y),
        'circle_correlations': create_circle_of_correlations(pca_results, feature_names)
    }
    
    # Save plots as HTML
    for name, fig in plots.items():
        filename = get_output_path(f"plotly_{name}.html")
        fig.write_html(filename)
        print(f"Saved: {filename}")
    
    # Save PCA results as JSON for validation
    results_for_json = {
        'explained_variance_ratio': pca_results['explained_variance_ratio'].tolist(),
        'cumulative_variance': pca_results['cumulative_variance'].tolist(),
        'n_components': pca_results['n_components'],
        'loadings_shape': pca_results['loadings'].shape,
        'scores_shape': pca_results['scores'].shape
    }
    
    with open(get_output_path('pca_results.json'), 'w') as f:
        json.dump(results_for_json, f, indent=2)
    print(f"\nSaved: {get_output_path('pca_results.json')}")
    
    return plots, pca_results


if __name__ == "__main__":
    plots, pca_results = save_all_plots()
    print("\n✅ All reference plots created successfully!")
    print(f"Open the HTML files in the '{OUTPUT_DIR}' folder to view interactive visualizations.")