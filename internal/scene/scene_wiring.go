package scene

// Wiring API — exported accessors letting the serializer stage (internal/aep,
// which holds the concrete back-ref impls + parse/mutate code) populate and
// reach the unexported serializer fields on scene runtime types without naming
// them directly. These exist purely to bridge the scene/serializer package
// boundary; they are not part of the user-facing API surface (the JSON export
// and the Set*/getter methods never touch them).

// --- back-ref interface fields (get + set) ---------------------------------

func ProjectBack(p *Project) ProjectWriter                                { return p.back }
func SetProjectBack(p *Project, w ProjectWriter)                          { p.back = w }
func CompositionBack(c *Composition) CompositionWriter                    { return c.back }
func SetCompositionBack(c *Composition, w CompositionWriter)              { c.back = w }
func CompositionProj(c *Composition) *Project                             { return c.proj }
func SetCompositionProj(c *Composition, p *Project)                       { c.proj = p }
func LayerBack(l *Layer) LayerWriter                                      { return l.back }
func SetLayerBack(l *Layer, w LayerWriter)                                { l.back = w }
func FootageBack(f *Footage) FootageWriter                                { return f.back }
func SetFootageBack(f *Footage, w FootageWriter)                          { f.back = w }
func PropertyBack(p *Property) PropertyWriter                             { return p.back }
func SetPropertyBack(p *Property, w PropertyWriter)                       { p.back = w }
func KeyframeBack(k *Keyframe) KeyframeWriter                             { return k.back }
func SetKeyframeBack(k *Keyframe, w KeyframeWriter)                       { k.back = w }
func MarkerBack(m *Marker) MarkerWriter                                   { return m.back }
func SetMarkerBack(m *Marker, w MarkerWriter)                             { m.back = w }
func MaskBack(m *Mask) MaskWriter                                         { return m.back }
func SetMaskBack(m *Mask, w MaskWriter)                                   { m.back = w }
func RenderQueueBack(rq *RenderQueue) RenderQueueWriter                   { return rq.back }
func SetRenderQueueBack(rq *RenderQueue, w RenderQueueWriter)             { rq.back = w }
func RenderQueueItemBack(it *RenderQueueItem) RenderQueueItemWriter       { return it.back }
func SetRenderQueueItemBack(it *RenderQueueItem, w RenderQueueItemWriter) { it.back = w }
func OutputModuleBack(om *OutputModule) OutputModuleWriter                { return om.back }
func SetOutputModuleBack(om *OutputModule, w OutputModuleWriter)          { om.back = w }
func PropertyGroupBack(g *AEPropertyGroup) PropertyGroupWriter            { return g.back }
func SetPropertyGroupBack(g *AEPropertyGroup, w PropertyGroupWriter)      { g.back = w }

// --- Marker serializer fields ----------------------------------------------

func MarkerSetList(m *Marker) MarkerSetRef         { return m.list }
func SetMarkerSetList(m *Marker, ref MarkerSetRef) { m.list = ref }
func MarkerLdatOffset(m *Marker) int               { return m.ldatOffset }
func SetMarkerLdatOffset(m *Marker, off int)       { m.ldatOffset = off }
func SetMarkerRates(m *Marker, tick, fps float64)  { m.tickRate, m.compFps = tick, fps }
func MarkerCompFps(m *Marker) float64              { return m.compFps }

// --- AEPropertyGroup layer backpointer -------------------------------------

func PropertyGroupLayer(g *AEPropertyGroup) *Layer       { return g.layer }
func SetPropertyGroupLayer(g *AEPropertyGroup, l *Layer) { g.layer = l }

// --- Render-queue scene-owned settings buffers -----------------------------

func RenderQueueItemSettings(it *RenderQueueItem) []byte       { return it.settingsBlock }
func SetRenderQueueItemSettings(it *RenderQueueItem, b []byte) { it.settingsBlock = b }
func OutputModuleSettingsBlock(om *OutputModule) []byte        { return om.settingsBlock }
func SetOutputModuleSettingsBlock(om *OutputModule, b []byte)  { om.settingsBlock = b }
func OutputModuleRoouData(om *OutputModule) []byte             { return om.roouData }
func SetOutputModuleRoouData(om *OutputModule, b []byte)       { om.roouData = b }

// --- Project structural-mutation fields ------------------------------------

func ProjectNextItemID(p *Project) uint32        { return p.nextItemID }
func SetProjectNextItemID(p *Project, id uint32) { p.nextItemID = id }
func ProjectTarget(p *Project) AETarget          { return p.target }
func SetProjectTarget(p *Project, t AETarget)    { p.target = t }

// --- Layer runtime-graph + owning-comp fields ------------------------------

func LayerComp(l *Layer) *Composition                    { return l.comp }
func SetLayerComp(l *Layer, c *Composition)              { l.comp = c }
func LayerPropertyTree(l *Layer) *AEPropertyGroup        { return l.propertyTree }
func SetLayerPropertyTree(l *Layer, g *AEPropertyGroup)  { l.propertyTree = g }
func LayerShapeRootGroup(l *Layer) *VectorGroup          { return l.shapeRootGroup }
func SetLayerShapeRootGroup(l *Layer, g *VectorGroup)    { l.shapeRootGroup = g }
func LayerShapeTransform(l *Layer) *LayerTransform       { return l.shapeTransform }
func SetLayerShapeTransform(l *Layer, t *LayerTransform) { l.shapeTransform = t }
func LayerShapeDirty(l *Layer) bool                      { return l.shapeDirty }
func SetLayerShapeDirty(l *Layer, v bool)                { l.shapeDirty = v }

// --- Guide scene-owned block ------------------------------------------------

func GuideBlock(g *Guide) []byte       { return g.block }
func SetGuideBlock(g *Guide, b []byte) { g.block = b }

// --- AEPropertyGroup tree wiring --------------------------------------------

func PropertyGroupParent(g *AEPropertyGroup) *AEPropertyGroup       { return g.parent }
func SetPropertyGroupParent(g *AEPropertyGroup, p *AEPropertyGroup) { g.parent = p }

func PropertyParentTreeGroup(p *Property) *AEPropertyGroup       { return p.parentTreeGroup }
func SetPropertyParentTreeGroup(p *Property, g *AEPropertyGroup) { p.parentTreeGroup = g }

// --- Shape-node scalar value fields (serializer-stage hydrate raw set) -------

func SetRectNodeDirection(r *RectNode, d ShapeDirection)               { r.direction = d }
func SetEllipseNodeDirection(e *EllipseNode, d ShapeDirection)         { e.direction = d }
func SetFillNodeBlendMode(f *FillNode, m ShapeBlendMode)               { f.blendMode = m }
func SetFillNodeCompositeOrder(f *FillNode, o ShapeCompositeOrder)     { f.compositeOrder = o }
func SetFillNodeFillRule(f *FillNode, r FillRule)                      { f.fillRule = r }
func SetGradientFillNodeGradient(n *GradientFillNode, g *Gradient)      { n.gradient = g }
func SetGradientStrokeNodeGradient(n *GradientStrokeNode, g *Gradient)  { n.gradient = g }
func SetStrokeNodeLineCap(s *StrokeNode, c StrokeLineCap)              { s.lineCap = c }
func SetStrokeNodeLineJoin(s *StrokeNode, j StrokeLineJoin)            { s.lineJoin = j }
func SetStrokeNodeMiterLimit(s *StrokeNode, v float64)                 { s.miterLimit = v }
func SetStrokeNodeBlendMode(s *StrokeNode, m ShapeBlendMode)           { s.blendMode = m }
func SetStrokeNodeCompositeOrder(s *StrokeNode, o ShapeCompositeOrder) { s.compositeOrder = o }

// --- LayerTransform component streams ---------------------------------------

func LayerTransformAnchorPoint(t *LayerTransform) *PropertyStream[[2]float64] { return t.anchorPoint }
func LayerTransformPosition(t *LayerTransform) *PropertyStream[[2]float64]    { return t.position }
func LayerTransformScale(t *LayerTransform) *PropertyStream[[2]float64]       { return t.scale }
func LayerTransformRotation(t *LayerTransform) *PropertyStream[float64]       { return t.rotation }
func LayerTransformOpacity(t *LayerTransform) *PropertyStream[float64]        { return t.opacity }
