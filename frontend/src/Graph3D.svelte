<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import ForceGraph3D from '3d-force-graph';
  import * as THREE from 'three';
  import type { main } from '../wailsjs/go/models';

  interface Props {
    graph: main.GraphDTO;
    learnerStates: Record<string, main.LearnerMasteryDTO>;
    selectedGuideId: string;
    onNodeSelect: (id: string) => void;
    onClose?: () => void;
  }

  let { graph, learnerStates, selectedGuideId, onNodeSelect, onClose }: Props = $props();

  let container: HTMLDivElement;
  let viewportElement: HTMLDivElement;
  let graphInstance: any = null;
  let activeHoveredNode = $state<string | null>(null);
  let isFullscreen = $state(false);
  let legendCollapsed = $state(false);
  let resizeObserver: ResizeObserver | null = null;

  // Keep 3D node colors synchronized when selectedGuideId or learnerStates change
  $effect(() => {
    const _sel = selectedGuideId;
    const _states = learnerStates;
    if (graphInstance && _sel !== undefined && _states !== undefined) {
      graphInstance.nodeColor(graphInstance.nodeColor());
    }
  });

  function resetCamera() {
    if (!graphInstance) return;
    graphInstance.cameraPosition({ x: 0, y: 190, z: 440 }, { x: 0, y: 0, z: 0 }, 800);
  }

  // Layer elevation spacing along Y-axis
  function getScopeHeightMap(scopes: string[]) {
    const spacing = 150;
    const map = new Map<string, number>();
    const count = scopes.length;
    scopes.forEach((scope, index) => {
      // Level 0 (definition) at bottom, lessons at top
      const y = (index - (count - 1) / 2) * spacing;
      map.set(scope, y);
    });
    return map;
  }

  function getMastery(guideId: string): number {
    const key = `${guideId}:ALL`;
    if (learnerStates[key]) {
      return Math.round(learnerStates[key].p_mastery * 100);
    }
    for (const [k, v] of Object.entries(learnerStates)) {
      if (k.startsWith(`${guideId}:`)) {
        return Math.round(v.p_mastery * 100);
      }
    }
    return 10;
  }

  // Grayscale Luminosity Encoding for Mastery
  function getNodeColor(guide: main.GuideSummaryDTO): string {
    if (guide.id === selectedGuideId) {
      return '#ffffff'; // Pure luminous white
    }
    const mastery = getMastery(guide.id);
    if (mastery >= 90) return '#f4f4f5'; // 90%+ Mastered: Bright White/Zinc
    if (mastery >= 50) return '#a1a1aa'; // 50-89% Proficient: Medium Silver
    if (mastery >= 20) return '#52525b'; // 20-49% Familiar: Slate Zinc
    return '#27272a'; // <20% Unlearned: Dim Charcoal
  }

  // Radial outward curve rotation for 360° symmetry
  function getRadialCurveRotation(link: any): number {
    const start = link.source;
    const end = link.target;
    if (!start || !end || typeof start.x !== 'number' || typeof end.x !== 'number') {
      return 0;
    }

    const vLine = new THREE.Vector3(
      end.x - start.x,
      (end.y || 0) - (start.y || 0),
      end.z - start.z
    );
    const len = vLine.length();
    if (len < 1e-4) return 0;

    const vAxis = vLine.clone().normalize();
    const dx = end.x - start.x;
    const dy = (end.y || 0) - (start.y || 0);
    const vBase = vLine
      .clone()
      .cross(dx !== 0 || dy !== 0 ? new THREE.Vector3(0, 0, 1) : new THREE.Vector3(0, 1, 0))
      .normalize();

    const mx = (start.x + end.x) / 2;
    const mz = (start.z + end.z) / 2;
    const rLen = Math.hypot(mx, mz);
    if (rLen < 1e-4) return 0;

    const rOut = new THREE.Vector3(mx / rLen, 0, mz / rLen);
    const rPerp = rOut.clone().sub(vAxis.clone().multiplyScalar(rOut.dot(vAxis)));
    if (rPerp.length() < 1e-4) return 0;
    rPerp.normalize();

    const cross = new THREE.Vector3().crossVectors(vBase, rPerp);
    const dot = vBase.dot(rPerp);
    return Math.atan2(vAxis.dot(cross), dot);
  }

  // Monochromatic Minimalist Radar Rings
  function createMonochromeHaloDisc(radius: number, yElevation: number): THREE.Group {
    const group = new THREE.Group();
    group.position.y = yElevation;

    // Semi-transparent base floor disc
    const discGeo = new THREE.CircleGeometry(radius, 64);
    const discMat = new THREE.MeshBasicMaterial({
      color: 0x09090b,
      transparent: true,
      opacity: 0.5,
      depthWrite: false,
      side: THREE.DoubleSide
    });
    const discMesh = new THREE.Mesh(discGeo, discMat);
    discMesh.rotation.x = -Math.PI / 2;
    group.add(discMesh);

    const makeRing = (r: number, color: number, opacity: number, dashed = false) => {
      const segments = 64;
      const points: THREE.Vector3[] = [];
      for (let i = 0; i <= segments; i++) {
        const theta = (i / segments) * Math.PI * 2;
        points.push(new THREE.Vector3(Math.cos(theta) * r, 0, Math.sin(theta) * r));
      }
      const geo = new THREE.BufferGeometry().setFromPoints(points);
      const mat = dashed
        ? new THREE.LineDashedMaterial({ color, opacity, transparent: true, dashSize: 6, gapSize: 4 })
        : new THREE.LineBasicMaterial({ color, opacity, transparent: true });
      const line = new THREE.Line(geo, mat);
      if (dashed) line.computeLineDistances();
      return line;
    };

    group.add(makeRing(radius, 0x3f3f46, 0.4));           // Outer border
    group.add(makeRing(radius * 0.70, 0x27272a, 0.5));      // Mid ring
    group.add(makeRing(radius * 0.40, 0x52525b, 0.25, true)); // Inner dashed ring
    group.add(makeRing(radius * 0.15, 0x27272a, 0.4));      // Core ring

    // Hairline crosshair spokes
    const spokePoints = [
      new THREE.Vector3(-radius, 0, 0), new THREE.Vector3(radius, 0, 0),
      new THREE.Vector3(0, 0, -radius), new THREE.Vector3(0, 0, radius)
    ];
    const spokeGeo = new THREE.BufferGeometry().setFromPoints(spokePoints);
    const spokeMat = new THREE.LineBasicMaterial({ color: 0x27272a, transparent: true, opacity: 0.35 });
    group.add(new THREE.LineSegments(spokeGeo, spokeMat));

    return group;
  }

  function toggleFullscreen() {
    if (!viewportElement) return;
    if (!document.fullscreenElement) {
      viewportElement.requestFullscreen().then(() => isFullscreen = true).catch(() => {});
    } else {
      document.exitFullscreen().then(() => isFullscreen = false).catch(() => {});
    }
  }

  function handleFullscreenChange() {
    isFullscreen = !!document.fullscreenElement;
    if (graphInstance && container) {
      setTimeout(() => {
        graphInstance.width(container.clientWidth).height(container.clientHeight);
      }, 50);
    }
  }

  onMount(() => {
    if (!container || !graph) return;

    document.addEventListener('fullscreenchange', handleFullscreenChange);

    const scopeYMap = getScopeHeightMap(graph.scopes);

    const nodes = graph.guides.map(g => ({
      id: g.id,
      name: g.id,
      scope: g.scope,
      fy: scopeYMap.get(g.scope) ?? 0,
      val: g.scope === 'lesson' ? 8 : g.scope === 'explanation' ? 6 : g.scope === 'description' ? 5 : 4,
      mastery: getMastery(g.id)
    }));

    const links: any[] = [];
    const validGuideIds = new Set(graph.guides.map(g => g.id));

    graph.guides.forEach(g => {
      // 1. Horizontal Prerequisite Edges (Solid Crisp Lines)
      g.prerequisites?.forEach(p => {
        if (validGuideIds.has(p)) {
          links.push({
            source: p,
            target: g.id,
            type: 'prerequisite',
            name: `${p} → ${g.id} (Prerequisite)`
          });
        }
      });

      // 2. Vertical Subguide Edges (Dashed Lines toward Encompassing Guide)
      g.sub_guides?.forEach(sg => {
        if (validGuideIds.has(sg.guide)) {
          links.push({
            source: sg.guide,
            target: g.id,
            type: 'subguide',
            adherence: sg.adherence,
            segment: sg.segment,
            name: `${sg.guide} ⤴ ${g.id} (Subguide [${sg.adherence}])`
          });
        }
      });
    });

    // Monochrome Arrowhead Mesh for traveling subguide pulses
    const subguideArrowGeometry = new THREE.ConeGeometry(3.2, 8.0, 10);
    subguideArrowGeometry.rotateX(Math.PI / 2);
    const subguideArrowMaterial = new THREE.MeshBasicMaterial({
      color: 0xffffff,
      transparent: true,
      opacity: 0.95
    });
    const subguideArrowMesh = new THREE.Mesh(subguideArrowGeometry, subguideArrowMaterial);

    const elem = ForceGraph3D()(container)
      .backgroundColor('#000000') // Pure void black
      .showNavInfo(false)
      .nodeLabel((n: any) => `
        <div style="background: rgba(18,18,21,0.95); padding: 7px 11px; border-radius: 5px; border: 1px solid #3f3f46; font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 11px; color: #ffffff; box-shadow: 0 4px 14px rgba(0,0,0,0.7);">
          <div style="font-weight: 600; font-size: 12px; color: #ffffff;">${n.id}</div>
          <div style="color: #71717a; font-size: 10px; margin: 2px 0;">Tier: <strong style="color: #d4d4d8;">${n.scope}</strong></div>
          <div style="color: #a1a1aa; font-weight: 500;">Mastery: <span style="color: #ffffff; font-weight: 600;">${n.mastery}%</span></div>
        </div>
      `)
      .nodeColor((n: any) => {
        const g = graph.guides.find(x => x.id === n.id);
        return g ? getNodeColor(g) : '#3f3f46';
      })
      .nodeResolution(16)
      // Solid white for prerequisites, ash-zinc for subguides
      .linkColor((l: any) => l.type === 'prerequisite' ? '#ffffff' : '#71717a')
      .linkWidth((l: any) => l.type === 'prerequisite' ? 1.6 : 1.0)
      .linkCurvature((l: any) => l.type === 'subguide' ? 0.22 : 0.0)
      .linkCurveRotation((l: any) => l.type === 'subguide' ? getRadialCurveRotation(l) : 0)
      // Arrowhead at target
      .linkDirectionalArrowLength((l: any) => l.type === 'subguide' ? 7 : 5)
      .linkDirectionalArrowRelPos(1)
      .linkDirectionalArrowColor((l: any) => l.type === 'subguide' ? '#ffffff' : '#e4e4e7')
      // Traveling white chevrons
      .linkDirectionalParticles((l: any) => l.type === 'subguide' ? 3 : 0)
      .linkDirectionalParticleSpeed(0.007)
      .linkDirectionalParticleWidth(4.0)
      .linkDirectionalParticleThreeObject((l: any) => l.type === 'subguide' ? subguideArrowMesh.clone() : null)
      .onNodeClick((node: any) => {
        if (node?.id) {
          onNodeSelect(node.id);
        }
      })
      .onNodeHover((node: any) => {
        activeHoveredNode = node ? node.id : null;
      });

    // Custom Three.js link material for dashed rendering
    elem.linkMaterial((link: any) => {
      if (link.type === 'subguide') {
        return new THREE.LineDashedMaterial({
          color: 0x71717a,
          dashSize: 5,
          gapSize: 3,
          transparent: true,
          opacity: 0.6
        });
      }
      return new THREE.LineBasicMaterial({
        color: 0xffffff,
        transparent: true,
        opacity: 0.85
      });
    });

    // Add Monochromatic Halo Rings to Scene
    const scene = elem.scene();
    const haloRadius = 260;
    graph.scopes.forEach((scope, i) => {
      const y = scopeYMap.get(scope) ?? 0;
      const haloDisc = createMonochromeHaloDisc(haloRadius, y);
      scene.add(haloDisc);
    });

    elem.graphData({ nodes, links });
    elem.cameraPosition({ x: 0, y: 190, z: 440 }, { x: 0, y: 0, z: 0 }, 1000);

    resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        if (elem && width > 0 && height > 0) {
          elem.width(width).height(height);
        }
      }
    });
    resizeObserver.observe(container);

    graphInstance = elem;
  });

  onDestroy(() => {
    document.removeEventListener('fullscreenchange', handleFullscreenChange);
    if (resizeObserver) {
      resizeObserver.disconnect();
    }
    if (graphInstance && graphInstance._destructor) {
      graphInstance._destructor();
    }
  });
</script>

<div bind:this={viewportElement} class="graph3d-viewport {isFullscreen ? 'is-fullscreen' : ''}">
  <div bind:this={container} class="canvas-container"></div>

  <!-- Legend & Control HUD Overlay -->
  <div class="overlay-legend">
    <div class="legend-header">
      <span class="legend-title">3D Topology</span>
      <div class="legend-actions">
        <button class="btn-icon" title="Reset Camera Orbit" onclick={resetCamera}>
          ⟲ Reset
        </button>
        <button class="btn-icon" title="Toggle Fullscreen" onclick={toggleFullscreen}>
          {isFullscreen ? '⤓ Exit' : '⛶ Full'}
        </button>
        <button
          class="btn-icon"
          title={legendCollapsed ? 'Expand Legend' : 'Collapse Legend'}
          onclick={() => legendCollapsed = !legendCollapsed}
        >
          {legendCollapsed ? '▾' : '▴'}
        </button>
        {#if onClose}
          <button class="btn-close" title="Switch to Author Perspective" onclick={onClose}>✕</button>
        {/if}
      </div>
    </div>

    {#if !legendCollapsed}
      <div class="legend-scopes">
        <span class="legend-subtitle">Scope Tiers (Elevation):</span>
        {#each graph.scopes as sc, idx}
          <div class="scope-legend-row">
            <span class="layer-badge">T{idx}</span>
            <span class="scope-name">{sc}</span>
          </div>
        {/each}
      </div>

      <div class="legend-edges">
        <span class="legend-subtitle">Relations:</span>
        <div class="edge-row">
          <span class="edge-line solid-white"></span>
          <span>Solid White: Prerequisite (DAG)</span>
        </div>
        <div class="edge-row">
          <span class="edge-line dashed-zinc"></span>
          <span>Dashed Ash: Subguide ⤴ Parent</span>
        </div>
        <div class="edge-row">
          <span class="edge-dot white-dot"></span>
          <span>White Chevron: Upstream Flow</span>
        </div>
      </div>

      <div class="legend-hint">
        Orbit: Left-drag · Pan: Right-drag · Zoom: Scroll<br/>
        Click any node to inspect in drawer
      </div>
    {/if}
  </div>

  {#if activeHoveredNode}
    <div class="hover-chip">
      Node: <strong>{activeHoveredNode}</strong>
    </div>
  {/if}
</div>

<style>
  .graph3d-viewport {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    background: #000000;
  }
  .graph3d-viewport.is-fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 9999;
  }

  .canvas-container {
    width: 100%;
    height: 100%;
  }

  .overlay-legend {
    position: absolute;
    top: 14px;
    left: 14px;
    background: rgba(9, 9, 11, 0.92);
    backdrop-filter: blur(12px);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 12px;
    width: 240px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    font-size: 11px;
    color: var(--text-main);
    z-index: 10;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.8);
  }

  .legend-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 7px;
  }
  .legend-title {
    font-weight: 600;
    font-size: 12px;
    color: var(--text-main);
    letter-spacing: -0.2px;
  }
  .legend-actions {
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .btn-icon {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-muted);
    font-size: 10px;
    padding: 2px 6px;
    cursor: pointer;
    transition: all 0.1s ease;
  }
  .btn-icon:hover {
    color: var(--text-main);
    border-color: var(--border-focus);
  }
  .btn-close {
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 12px;
    cursor: pointer;
    padding: 0 3px;
  }
  .btn-close:hover {
    color: var(--text-main);
  }

  .legend-scopes, .legend-edges {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .legend-subtitle {
    font-weight: 600;
    text-transform: uppercase;
    font-size: 9px;
    color: var(--text-dim);
    letter-spacing: 0.5px;
  }
  .scope-legend-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .layer-badge {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 1px 4px;
    border-radius: 3px;
    font-family: monospace;
    font-size: 9px;
    color: var(--text-muted);
  }
  .scope-name {
    font-weight: 500;
    text-transform: capitalize;
    color: var(--text-muted);
  }

  .edge-row {
    display: flex;
    align-items: center;
    gap: 7px;
    color: var(--text-muted);
  }
  .edge-line {
    width: 20px;
    height: 2px;
    border-radius: 1px;
  }
  .edge-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    margin: 0 7px;
  }
  .solid-white {
    background: #ffffff;
  }
  .dashed-zinc {
    background: repeating-linear-gradient(
      90deg,
      #71717a,
      #71717a 3px,
      transparent 3px,
      transparent 6px
    );
  }
  .white-dot {
    background: #ffffff;
    box-shadow: 0 0 5px #ffffff;
  }

  .legend-hint {
    border-top: 1px solid var(--border-color);
    padding-top: 7px;
    color: var(--text-dim);
    line-height: 1.4;
    font-size: 9px;
  }

  .hover-chip {
    position: absolute;
    bottom: 16px;
    left: 50%;
    transform: translateX(-50%);
    background: rgba(18, 18, 21, 0.95);
    border: 1px solid var(--border-color);
    padding: 4px 12px;
    border-radius: 16px;
    font-size: 11px;
    color: var(--text-main);
    z-index: 10;
  }
</style>
