package v2

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Layr-Labs/eigenda/core"
	corev2 "github.com/Layr-Labs/eigenda/core/v2"
	"github.com/Layr-Labs/eigenda/disperser/common/semver"
	disperserv2 "github.com/Layr-Labs/eigenda/disperser/common/v2"
	"github.com/Layr-Labs/eigenda/disperser/common/v2/blobstore"
	"github.com/Layr-Labs/eigenda/disperser/dataapi"
	docsv2 "github.com/Layr-Labs/eigenda/disperser/dataapi/docs/v2"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

var errNotFound = errors.New("not found")

const (
	maxBlobAge = 14 * 24 * time.Hour

	// The max number of blobs to return from blob feed API, regardless of the time
	// range or "limit" param.
	maxNumBlobsPerBlobFeedResponse = 1000

	// The max number of batches to return from batch feed API, regardless of the time
	// range or "limit" param.
	maxNumBatchesPerBatchFeedResponse = 1000

	cacheControlParam       = "Cache-Control"
	maxFeedBlobAge          = 300 // this is completely static
	maxOperatorsStakeAge    = 300 // not expect the stake change to happen frequently
	maxOperatorResponseAge  = 300 // this is completely static
	maxOperatorPortCheckAge = 60
	maxMetricAge            = 10
	maxThroughputAge        = 10
)

type (
	ErrorResponse struct {
		Error string `json:"error"`
	}

	SignedBatch struct {
		BatchHeader *corev2.BatchHeader `json:"batch_header"`
		Attestation *corev2.Attestation `json:"attestation"`
	}

	BlobResponse struct {
		BlobKey       string             `json:"blob_key"`
		BlobHeader    *corev2.BlobHeader `json:"blob_header"`
		Status        string             `json:"status"`
		DispersedAt   uint64             `json:"dispersed_at"`
		BlobSizeBytes uint64             `json:"blob_size_bytes"`
	}

	BlobCertificateResponse struct {
		Certificate *corev2.BlobCertificate `json:"blob_certificate"`
	}

	BlobInclusionInfoResponse struct {
		InclusionInfo *corev2.BlobInclusionInfo `json:"blob_inclusion_info"`
	}

	BlobInfo struct {
		BlobKey      string                    `json:"blob_key"`
		BlobMetadata *disperserv2.BlobMetadata `json:"blob_metadata"`
	}
	BlobFeedResponse struct {
		Blobs           []BlobInfo `json:"blobs"`
		PaginationToken string     `json:"pagination_token"`
	}

	BatchResponse struct {
		BatchHeaderHash    string                      `json:"batch_header_hash"`
		SignedBatch        *SignedBatch                `json:"signed_batch"`
		BlobInclusionInfos []*corev2.BlobInclusionInfo `json:"blob_inclusion_infos"`
	}

	BatchInfo struct {
		BatchHeaderHash         string                  `json:"batch_header_hash"`
		BatchHeader             *corev2.BatchHeader     `json:"batch_header"`
		AttestedAt              uint64                  `json:"attested_at"`
		AggregatedSignature     *core.Signature         `json:"aggregated_signature"`
		QuorumNumbers           []core.QuorumID         `json:"quorum_numbers"`
		QuorumSignedPercentages map[core.QuorumID]uint8 `json:"quorum_signed_percentages"`
	}
	BatchFeedResponse struct {
		Batches []*BatchInfo `json:"batches"`
	}

	MetricSummary struct {
		AvgThroughput float64 `json:"avg_throughput"`
	}

	OperatorStake struct {
		QuorumId        string  `json:"quorum_id"`
		OperatorId      string  `json:"operator_id"`
		StakePercentage float64 `json:"stake_percentage"`
		Rank            int     `json:"rank"`
	}

	OperatorsStakeResponse struct {
		StakeRankedOperators map[string][]*OperatorStake `json:"stake_ranked_operators"`
	}

	// Operators' responses for a batch
	OperatorDispersalResponses struct {
		Responses []*corev2.DispersalResponse `json:"operator_dispersal_responses"`
	}

	OperatorPortCheckResponse struct {
		OperatorId        string `json:"operator_id"`
		DispersalSocket   string `json:"dispersal_socket"`
		DispersalOnline   bool   `json:"dispersal_online"`
		DispersalStatus   string `json:"dispersal_status"`
		RetrievalSocket   string `json:"retrieval_socket"`
		RetrievalOnline   bool   `json:"retrieval_online"`
		RetrievalStatus   string `json:"retrieval_status"`
		V2DispersalSocket string `json:"v2_dispersal_socket"`
		V2DispersalOnline bool   `json:"v2_dispersal_online"`
		V2DispersalStatus string `json:"v2_dispersal_status"`
	}

	SemverReportResponse struct {
		Semver map[string]*semver.SemverMetrics `json:"semver"`
	}

	Metric struct {
		Throughput float64 `json:"throughput"`
		CostInGas  float64 `json:"cost_in_gas"`
		// deprecated: use TotalStakePerQuorum instead. Remove when the frontend is updated.
		TotalStake          *big.Int                   `json:"total_stake"`
		TotalStakePerQuorum map[core.QuorumID]*big.Int `json:"total_stake_per_quorum"`
	}

	Throughput struct {
		Throughput float64 `json:"throughput"`
		Timestamp  uint64  `json:"timestamp"`
	}
)

type ServerV2 struct {
	serverMode   string
	socketAddr   string
	allowOrigins []string
	logger       logging.Logger

	blobMetadataStore *blobstore.BlobMetadataStore
	subgraphClient    dataapi.SubgraphClient
	chainReader       core.Reader
	chainState        core.ChainState
	indexedChainState core.IndexedChainState
	promClient        dataapi.PrometheusClient
	metrics           *dataapi.Metrics

	operatorHandler *dataapi.OperatorHandler
	metricsHandler  *dataapi.MetricsHandler
}

func NewServerV2(
	config dataapi.Config,
	blobMetadataStore *blobstore.BlobMetadataStore,
	promClient dataapi.PrometheusClient,
	subgraphClient dataapi.SubgraphClient,
	chainReader core.Reader,
	chainState core.ChainState,
	indexedChainState core.IndexedChainState,
	logger logging.Logger,
	metrics *dataapi.Metrics,
) *ServerV2 {
	l := logger.With("component", "DataAPIServerV2")
	return &ServerV2{
		logger:            l,
		serverMode:        config.ServerMode,
		socketAddr:        config.SocketAddr,
		allowOrigins:      config.AllowOrigins,
		blobMetadataStore: blobMetadataStore,
		promClient:        promClient,
		subgraphClient:    subgraphClient,
		chainReader:       chainReader,
		chainState:        chainState,
		indexedChainState: indexedChainState,
		metrics:           metrics,
		operatorHandler:   dataapi.NewOperatorHandler(l, metrics, chainReader, chainState, indexedChainState, subgraphClient),
		metricsHandler:    dataapi.NewMetricsHandler(promClient),
	}
}

func (s *ServerV2) Start() error {
	if s.serverMode == gin.ReleaseMode {
		// optimize performance and disable debug features.
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	basePath := "/api/v2"
	docsv2.SwaggerInfoV2.BasePath = basePath
	docsv2.SwaggerInfoV2.Host = os.Getenv("SWAGGER_HOST")

	v2 := router.Group(basePath)
	{
		blobs := v2.Group("/blobs")
		{
			blobs.GET("/feed", s.FetchBlobFeedHandler)
			blobs.GET("/:blob_key", s.FetchBlobHandler)
			blobs.GET("/:blob_key/certificate", s.FetchBlobCertificateHandler)
			blobs.GET("/:blob_key/inclusion-info", s.FetchBlobInclusionInfoHandler)
		}
		batches := v2.Group("/batches")
		{
			batches.GET("/feed", s.FetchBatchFeedHandler)
			batches.GET("/:batch_header_hash", s.FetchBatchHandler)
		}
		operators := v2.Group("/operators")
		{
			operators.GET("/nonsigners", s.FetchNonSingers)
			operators.GET("/stake", s.FetchOperatorsStake)
			operators.GET("/nodeinfo", s.FetchOperatorsNodeInfo)
			operators.GET("/reachability", s.CheckOperatorsReachability)
			operators.GET("/response/:batch_header_hash", s.FetchOperatorsResponses)
		}
		metrics := v2.Group("/metrics")
		{
			metrics.GET("/summary", s.FetchMetricsSummaryHandler)
			metrics.GET("/timeseries/throughput", s.FetchMetricsThroughputTimeseriesHandler)
		}
		swagger := v2.Group("/swagger")
		{
			swagger.GET("/*any", ginswagger.WrapHandler(swaggerfiles.Handler, ginswagger.InstanceName("V2"), ginswagger.URL("/api/v2/swagger/doc.json")))

		}
	}

	router.GET("/", func(g *gin.Context) {
		g.JSON(http.StatusAccepted, gin.H{"status": "OK"})
	})

	router.Use(logger.SetLogger(
		logger.WithSkipPath([]string{"/"}),
	))

	config := cors.DefaultConfig()
	config.AllowOrigins = s.allowOrigins
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "HEAD", "OPTIONS"}

	if s.serverMode != gin.ReleaseMode {
		config.AllowOrigins = []string{"*"}
	}
	router.Use(cors.New(config))

	srv := &http.Server{
		Addr:              s.socketAddr,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errChan := run(s.logger, srv)
	return <-errChan
}

func errorResponse(c *gin.Context, err error) {
	_ = c.Error(err)
	var code int
	switch {
	case errors.Is(err, errNotFound):
		code = http.StatusNotFound
	default:
		code = http.StatusInternalServerError
	}
	c.JSON(code, ErrorResponse{
		Error: err.Error(),
	})
}

func invalidParamsErrorResponse(c *gin.Context, err error) {
	_ = c.Error(err)
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: err.Error(),
	})
}

func run(logger logging.Logger, httpServer *http.Server) <-chan error {
	errChan := make(chan error, 1)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		<-ctx.Done()

		logger.Info("shutdown signal received")

		defer func() {
			stop()
			close(errChan)
		}()

		if err := httpServer.Shutdown(context.Background()); err != nil {
			errChan <- err
		}
		logger.Info("shutdown completed")
	}()

	go func() {
		logger.Info("server v2 running", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

func (s *ServerV2) Shutdown() error {
	return nil
}

// FetchBlobFeedHandler godoc
//
//	@Summary	Fetch blob feed
//	@Tags		Blob
//	@Produce	json
//	@Param		end					query		string	false	"Fetch blobs up to the end time (ISO 8601 format: 2006-01-02T15:04:05Z) [default: now]"
//	@Param		interval			query		int		false	"Fetch blobs starting from an interval (in seconds) before the end time [default: 3600]"
//	@Param		pagination_token	query		string	false	"Fetch blobs starting from the pagination token (exclusively). Overrides the interval param if specified [default: empty]"
//	@Param		limit				query		int		false	"The maximum number of blobs to fetch. System max (1000) if limit <= 0 [default: 20; max: 1000]"
//	@Success	200					{object}	BlobFeedResponse
//	@Failure	400					{object}	ErrorResponse	"error: Bad request"
//	@Failure	404					{object}	ErrorResponse	"error: Not found"
//	@Failure	500					{object}	ErrorResponse	"error: Server error"
//	@Router		/blobs/feed [get]
func (s *ServerV2) FetchBlobFeedHandler(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("FetchBlobFeedHandler", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	var err error

	now := time.Now()
	oldestTime := now.Add(-maxBlobAge)

	endTime := now
	if c.Query("end") != "" {
		endTime, err = time.Parse("2006-01-02T15:04:05Z", c.Query("end"))
		if err != nil {
			s.metrics.IncrementInvalidArgRequestNum("FetchBlobFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("failed to parse end param: %w", err))
			return
		}
		if endTime.Before(oldestTime) {
			s.metrics.IncrementInvalidArgRequestNum("FetchBlobFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("end time cannot be more than 14 days ago from now, found: %s", c.Query("end")))
			return
		}
	}

	interval := 3600
	if c.Query("interval") != "" {
		interval, err = strconv.Atoi(c.Query("interval"))
		if err != nil {
			s.metrics.IncrementInvalidArgRequestNum("FetchBlobFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("failed to parse interval param: %w", err))
			return
		}
		if interval <= 0 {
			s.metrics.IncrementInvalidArgRequestNum("FetchBlobFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("interval must be greater than 0, found: %d", interval))
			return
		}
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchBlobFeedHandler")
		invalidParamsErrorResponse(c, fmt.Errorf("failed to parse limit param: %w", err))
		return
	}
	if limit <= 0 || limit > maxNumBlobsPerBlobFeedResponse {
		limit = maxNumBlobsPerBlobFeedResponse
	}

	paginationCursor := blobstore.BlobFeedCursor{
		RequestedAt: 0,
	}
	if c.Query("pagination_token") != "" {
		cursor, err := paginationCursor.FromCursorKey(c.Query("pagination_token"))
		if err != nil {
			s.metrics.IncrementInvalidArgRequestNum("FetchBlobFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("failed to parse the pagination token: %w", err))
			return
		}
		paginationCursor = *cursor
	}

	startTime := endTime.Add(-time.Duration(interval) * time.Second)
	startCursor := blobstore.BlobFeedCursor{
		RequestedAt: uint64(startTime.UnixNano()),
	}
	if startCursor.LessThan(&paginationCursor) {
		startCursor = paginationCursor
	}
	endCursor := blobstore.BlobFeedCursor{
		RequestedAt: uint64(endTime.UnixNano()),
	}

	blobs, paginationToken, err := s.blobMetadataStore.GetBlobMetadataByRequestedAt(c.Request.Context(), startCursor, endCursor, limit)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchBlobFeedHandler")
		errorResponse(c, fmt.Errorf("failed to fetch feed from blob metadata store: %w", err))
		return
	}

	token := ""
	if paginationToken != nil {
		token = paginationToken.ToCursorKey()
	}
	blobInfo := make([]BlobInfo, len(blobs))
	for i := 0; i < len(blobs); i++ {
		bk, err := blobs[i].BlobHeader.BlobKey()
		if err != nil {
			s.metrics.IncrementFailedRequestNum("FetchBlobFeedHandler")
			errorResponse(c, fmt.Errorf("failed to serialize blob key: %w", err))
			return
		}
		blobInfo[i].BlobKey = bk.Hex()
		blobInfo[i].BlobMetadata = blobs[i]
	}
	response := &BlobFeedResponse{
		Blobs:           blobInfo,
		PaginationToken: token,
	}
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxFeedBlobAge))
	s.metrics.IncrementSuccessfulRequestNum("FetchBlobFeedHandler")
	c.JSON(http.StatusOK, response)
}

// FetchBlobHandler godoc
//
//	@Summary	Fetch blob metadata by blob key
//	@Tags		Blob
//	@Produce	json
//	@Param		blob_key	path		string	true	"Blob key in hex string"
//	@Success	200			{object}	BlobResponse
//	@Failure	400			{object}	ErrorResponse	"error: Bad request"
//	@Failure	404			{object}	ErrorResponse	"error: Not found"
//	@Failure	500			{object}	ErrorResponse	"error: Server error"
//	@Router		/blobs/{blob_key} [get]
func (s *ServerV2) FetchBlobHandler(c *gin.Context) {
	start := time.Now()
	blobKey, err := corev2.HexToBlobKey(c.Param("blob_key"))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchBlob")
		errorResponse(c, err)
		return
	}
	metadata, err := s.blobMetadataStore.GetBlobMetadata(c.Request.Context(), blobKey)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchBlob")
		errorResponse(c, err)
		return
	}
	bk, err := metadata.BlobHeader.BlobKey()
	if err != nil || bk != blobKey {
		s.metrics.IncrementFailedRequestNum("FetchBlob")
		errorResponse(c, err)
		return
	}
	response := &BlobResponse{
		BlobKey:       bk.Hex(),
		BlobHeader:    metadata.BlobHeader,
		Status:        metadata.BlobStatus.String(),
		DispersedAt:   metadata.RequestedAt,
		BlobSizeBytes: metadata.BlobSize,
	}
	s.metrics.IncrementSuccessfulRequestNum("FetchBlob")
	s.metrics.ObserveLatency("FetchBlob", float64(time.Since(start).Milliseconds()))
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxFeedBlobAge))
	c.JSON(http.StatusOK, response)
}

// FetchBlobCertificateHandler godoc
//
//	@Summary	Fetch blob certificate by blob key v2
//	@Tags		Blob
//	@Produce	json
//	@Param		blob_key	path		string	true	"Blob key in hex string"
//	@Success	200			{object}	BlobCertificateResponse
//	@Failure	400			{object}	ErrorResponse	"error: Bad request"
//	@Failure	404			{object}	ErrorResponse	"error: Not found"
//	@Failure	500			{object}	ErrorResponse	"error: Server error"
//	@Router		/blobs/{blob_key}/certificate [get]
func (s *ServerV2) FetchBlobCertificateHandler(c *gin.Context) {
	start := time.Now()
	blobKey, err := corev2.HexToBlobKey(c.Param("blob_key"))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchBlobCertificate")
		errorResponse(c, err)
		return
	}
	cert, _, err := s.blobMetadataStore.GetBlobCertificate(c.Request.Context(), blobKey)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchBlobCertificate")
		errorResponse(c, err)
		return
	}
	response := &BlobCertificateResponse{
		Certificate: cert,
	}
	s.metrics.IncrementSuccessfulRequestNum("FetchBlobCertificate")
	s.metrics.ObserveLatency("FetchBlobCertificate", float64(time.Since(start).Milliseconds()))
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxFeedBlobAge))
	c.JSON(http.StatusOK, response)
}

// FetchBlobInclusionInfoHandler godoc
//
//	@Summary	Fetch blob inclusion info by blob key and batch header hash
//	@Tags		Blob
//	@Produce	json
//	@Param		blob_key			path		string	true	"Blob key in hex string"
//	@Param		batch_header_hash	path		string	true	"Batch header hash in hex string"
//
//	@Success	200					{object}	BlobInclusionInfoResponse
//	@Failure	400					{object}	ErrorResponse	"error: Bad request"
//	@Failure	404					{object}	ErrorResponse	"error: Not found"
//	@Failure	500					{object}	ErrorResponse	"error: Server error"
//	@Router		/blobs/{blob_key}/inclusion-info [get]
func (s *ServerV2) FetchBlobInclusionInfoHandler(c *gin.Context) {
	start := time.Now()
	blobKey, err := corev2.HexToBlobKey(c.Param("blob_key"))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchBlobInclusionInfo")
		errorResponse(c, err)
		return
	}
	batchHeaderHashHex := c.Query("batch_header_hash")
	batchHeaderHash, err := dataapi.ConvertHexadecimalToBytes([]byte(batchHeaderHashHex))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchBlobInclusionInfo")
		errorResponse(c, err)
		return
	}
	bvi, err := s.blobMetadataStore.GetBlobInclusionInfo(c.Request.Context(), blobKey, batchHeaderHash)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchBlobInclusionInfo")
		errorResponse(c, err)
		return
	}
	response := &BlobInclusionInfoResponse{
		InclusionInfo: bvi,
	}
	s.metrics.IncrementSuccessfulRequestNum("FetchBlobInclusionInfo")
	s.metrics.ObserveLatency("FetchBlobInclusionInfo", float64(time.Since(start).Milliseconds()))
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxFeedBlobAge))
	c.JSON(http.StatusOK, response)
}

// FetchBatchFeedHandler godoc
//
//	@Summary	Fetch batch feed
//	@Tags		Blob
//	@Produce	json
//	@Param		end			query		string	false	"Fetch batches up to the end time (ISO 8601 format: 2006-01-02T15:04:05Z) [default: now]"
//	@Param		interval	query		int		false	"Fetch batches starting from an interval (in seconds) before the end time [default: 3600]"
//	@Param		limit		query		int		false	"The maximum number of batches to fetch. System max (1000) if limit <= 0 [default: 20; max: 1000]"
//	@Success	200			{object}	BatchFeedResponse
//	@Failure	400			{object}	ErrorResponse	"error: Bad request"
//	@Failure	404			{object}	ErrorResponse	"error: Not found"
//	@Failure	500			{object}	ErrorResponse	"error: Server error"
//	@Router		/batches/feed [get]
func (s *ServerV2) FetchBatchFeedHandler(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("FetchBatchFeedHandler", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	var err error

	now := time.Now()
	oldestTime := now.Add(-maxBlobAge)

	endTime := now
	if c.Query("end") != "" {
		endTime, err = time.Parse("2006-01-02T15:04:05Z", c.Query("end"))
		if err != nil {
			s.metrics.IncrementInvalidArgRequestNum("FetchBatchFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("failed to parse end param: %w", err))
			return
		}
		if endTime.Before(oldestTime) {
			s.metrics.IncrementInvalidArgRequestNum("FetchBatchFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("end time cannot be more than 14 days ago from now, found: %s", c.Query("end")))
			return
		}
	}

	interval := 3600
	if c.Query("interval") != "" {
		interval, err = strconv.Atoi(c.Query("interval"))
		if err != nil {
			s.metrics.IncrementInvalidArgRequestNum("FetchBatchFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("failed to parse interval param: %w", err))
			return
		}
		if interval <= 0 {
			s.metrics.IncrementInvalidArgRequestNum("FetchBatchFeedHandler")
			invalidParamsErrorResponse(c, fmt.Errorf("interval must be greater than 0, found: %d", interval))
			return
		}
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchBatchFeedHandler")
		invalidParamsErrorResponse(c, fmt.Errorf("failed to parse limit param: %w", err))
		return
	}
	if limit <= 0 || limit > maxNumBatchesPerBatchFeedResponse {
		limit = maxNumBatchesPerBatchFeedResponse
	}

	startTime := endTime.Add(-time.Duration(interval) * time.Second)
	attestations, err := s.blobMetadataStore.GetAttestationByAttestedAt(c.Request.Context(), uint64(startTime.UnixNano())+1, uint64(endTime.UnixNano()), limit)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchBatchFeedHandler")
		errorResponse(c, fmt.Errorf("failed to fetch feed from blob metadata store: %w", err))
		return
	}

	batches := make([]*BatchInfo, len(attestations))
	for i, at := range attestations {
		batchHeaderHash, err := at.BatchHeader.Hash()
		if err != nil {
			s.metrics.IncrementFailedRequestNum("FetchBatchFeedHandler")
			errorResponse(c, fmt.Errorf("failed to compute batch header hash from batch header: %w", err))
			return
		}

		batches[i] = &BatchInfo{
			BatchHeaderHash:         hex.EncodeToString(batchHeaderHash[:]),
			BatchHeader:             at.BatchHeader,
			AttestedAt:              at.AttestedAt,
			AggregatedSignature:     at.Sigma,
			QuorumNumbers:           at.QuorumNumbers,
			QuorumSignedPercentages: at.QuorumResults,
		}
	}
	response := &BatchFeedResponse{
		Batches: batches,
	}
	s.metrics.IncrementSuccessfulRequestNum("FetchBatchFeedHandler")
	c.JSON(http.StatusOK, response)
}

// FetchBatchHandler godoc
//
//	@Summary	Fetch batch by the batch header hash
//	@Tags		Batch
//	@Produce	json
//	@Param		batch_header_hash	path		string	true	"Batch header hash in hex string"
//	@Success	200					{object}	BatchResponse
//	@Failure	400					{object}	ErrorResponse	"error: Bad request"
//	@Failure	404					{object}	ErrorResponse	"error: Not found"
//	@Failure	500					{object}	ErrorResponse	"error: Server error"
//	@Router		/batches/{batch_header_hash} [get]
func (s *ServerV2) FetchBatchHandler(c *gin.Context) {
	start := time.Now()
	batchHeaderHashHex := c.Param("batch_header_hash")
	batchHeaderHash, err := dataapi.ConvertHexadecimalToBytes([]byte(batchHeaderHashHex))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchBatch")
		errorResponse(c, errors.New("invalid batch header hash"))
		return
	}
	batchHeader, attestation, err := s.blobMetadataStore.GetSignedBatch(c.Request.Context(), batchHeaderHash)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchBatch")
		errorResponse(c, err)
		return
	}
	// TODO: support fetch of blob inclusion info
	batchResponse := &BatchResponse{
		BatchHeaderHash: batchHeaderHashHex,
		SignedBatch: &SignedBatch{
			BatchHeader: batchHeader,
			Attestation: attestation,
		},
	}
	s.metrics.IncrementSuccessfulRequestNum("FetchBatch")
	s.metrics.ObserveLatency("FetchBatch", float64(time.Since(start).Milliseconds()))
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxFeedBlobAge))
	c.JSON(http.StatusOK, batchResponse)
}

// FetchOperatorsStake godoc
//
//	@Summary	Operator stake distribution query
//	@Tags		Operators
//	@Produce	json
//	@Param		operator_id	query		string	false	"Operator ID in hex string [default: all operators if unspecified]"
//	@Success	200			{object}	OperatorsStakeResponse
//	@Failure	400			{object}	ErrorResponse	"error: Bad request"
//	@Failure	404			{object}	ErrorResponse	"error: Not found"
//	@Failure	500			{object}	ErrorResponse	"error: Server error"
//	@Router		/operators/stake [get]
func (s *ServerV2) FetchOperatorsStake(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("FetchOperatorsStake", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	operatorId := c.DefaultQuery("operator_id", "")
	s.logger.Info("getting operators stake distribution", "operatorId", operatorId)

	operatorsStakeResponse, err := s.operatorHandler.GetOperatorsStake(c.Request.Context(), operatorId)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchOperatorsStake")
		errorResponse(c, fmt.Errorf("failed to get operator stake - %s", err))
		return
	}

	s.metrics.IncrementSuccessfulRequestNum("FetchOperatorsStake")
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxOperatorsStakeAge))
	c.JSON(http.StatusOK, operatorsStakeResponse)
}

// FetchOperatorsNodeInfo godoc
//
//	@Summary	Active operator semver
//	@Tags		Operators
//	@Produce	json
//	@Success	200	{object}	SemverReportResponse
//	@Failure	500	{object}	ErrorResponse	"error: Server error"
//	@Router		/operators/nodeinfo [get]
func (s *ServerV2) FetchOperatorsNodeInfo(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("FetchOperatorsNodeInfo", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	report, err := s.operatorHandler.ScanOperatorsHostInfo(c.Request.Context())
	if err != nil {
		s.logger.Error("failed to scan operators host info", "error", err)
		s.metrics.IncrementFailedRequestNum("FetchOperatorsNodeInfo")
		errorResponse(c, err)
	}
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxOperatorPortCheckAge))
	c.JSON(http.StatusOK, report)
}

// FetchOperatorsResponses godoc
//
//	@Summary	Fetch operator attestation response for a batch
//	@Tags		Operators
//	@Produce	json
//	@Param		batch_header_hash	path		string	true	"Batch header hash in hex string"
//	@Param		operator_id			query		string	false	"Operator ID in hex string [default: all operators if unspecified]"
//	@Success	200					{object}	OperatorDispersalResponses
//	@Failure	400					{object}	ErrorResponse	"error: Bad request"
//	@Failure	404					{object}	ErrorResponse	"error: Not found"
//	@Failure	500					{object}	ErrorResponse	"error: Server error"
//	@Router		/operators/{batch_header_hash} [get]
func (s *ServerV2) FetchOperatorsResponses(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("FetchOperatorsResponses", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	batchHeaderHashHex := c.Param("batch_header_hash")
	batchHeaderHash, err := dataapi.ConvertHexadecimalToBytes([]byte(batchHeaderHashHex))
	if err != nil {
		s.metrics.IncrementInvalidArgRequestNum("FetchOperatorsResponses")
		errorResponse(c, errors.New("invalid batch header hash"))
		return
	}
	operatorIdStr := c.DefaultQuery("operator_id", "")

	operatorResponses := make([]*corev2.DispersalResponse, 0)
	if operatorIdStr == "" {
		res, err := s.blobMetadataStore.GetDispersalResponses(c.Request.Context(), batchHeaderHash)
		if err != nil {
			s.metrics.IncrementFailedRequestNum("FetchOperatorsResponses")
			errorResponse(c, err)
			return
		}
		operatorResponses = append(operatorResponses, res...)
	} else {
		operatorId, err := core.OperatorIDFromHex(operatorIdStr)
		if err != nil {
			s.metrics.IncrementInvalidArgRequestNum("FetchOperatorsResponses")
			errorResponse(c, errors.New("invalid operatorId"))
			return
		}

		res, err := s.blobMetadataStore.GetDispersalResponse(c.Request.Context(), batchHeaderHash, operatorId)
		if err != nil {
			s.metrics.IncrementFailedRequestNum("FetchOperatorsResponses")
			errorResponse(c, err)
			return
		}
		operatorResponses = append(operatorResponses, res)
	}
	response := &OperatorDispersalResponses{
		Responses: operatorResponses,
	}
	s.metrics.IncrementSuccessfulRequestNum("FetchOperatorsResponses")
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxOperatorResponseAge))
	c.JSON(http.StatusOK, response)
}

// CheckOperatorsReachability godoc
//
//	@Summary	Operator node reachability check
//	@Tags		Operators
//	@Produce	json
//	@Param		operator_id	query		string	false	"Operator ID in hex string [default: all operators if unspecified]"
//	@Success	200			{object}	OperatorPortCheckResponse
//	@Failure	400			{object}	ErrorResponse	"error: Bad request"
//	@Failure	404			{object}	ErrorResponse	"error: Not found"
//	@Failure	500			{object}	ErrorResponse	"error: Server error"
//	@Router		/operators/reachability [get]
func (s *ServerV2) CheckOperatorsReachability(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("OperatorPortCheck", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	operatorId := c.DefaultQuery("operator_id", "")
	s.logger.Info("checking operator ports", "operatorId", operatorId)
	portCheckResponse, err := s.operatorHandler.ProbeOperatorHosts(c.Request.Context(), operatorId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			err = errNotFound
			s.logger.Warn("operator not found", "operatorId", operatorId)
			s.metrics.IncrementNotFoundRequestNum("OperatorPortCheck")
		} else {
			s.logger.Error("operator port check failed", "error", err)
			s.metrics.IncrementFailedRequestNum("OperatorPortCheck")
		}
		errorResponse(c, err)
		return
	}
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxOperatorPortCheckAge))
	c.JSON(http.StatusOK, portCheckResponse)
}

func (s *ServerV2) FetchNonSingers(c *gin.Context) {
	errorResponse(c, errors.New("FetchNonSingers unimplemented"))
}

// FetchMetricsSummaryHandler godoc
//
//	@Summary	Fetch metrics summary
//	@Tags		Metrics
//	@Produce	json
//	@Param		start	query		int	false	"Start unix timestamp [default: 1 hour ago]"
//	@Param		end		query		int	false	"End unix timestamp [default: unix time now]"
//	@Success	200		{object}	Metric
//	@Failure	400		{object}	ErrorResponse	"error: Bad request"
//	@Failure	404		{object}	ErrorResponse	"error: Not found"
//	@Failure	500		{object}	ErrorResponse	"error: Server error"
//	@Router		/metrics/summary  [get]
func (s *ServerV2) FetchMetricsSummaryHandler(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("FetchMetricsSummary", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	now := time.Now()
	start, err := strconv.ParseInt(c.DefaultQuery("start", "0"), 10, 64)
	if err != nil || start == 0 {
		start = now.Add(-time.Hour * 1).Unix()
	}

	end, err := strconv.ParseInt(c.DefaultQuery("end", "0"), 10, 64)
	if err != nil || end == 0 {
		end = now.Unix()
	}

	avgThroughput, err := s.metricsHandler.GetAvgThroughput(c.Request.Context(), start, end)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchMetricsSummary")
		errorResponse(c, err)
		return
	}

	metricSummary := &MetricSummary{
		AvgThroughput: avgThroughput,
	}

	s.metrics.IncrementSuccessfulRequestNum("FetchMetricsSummary")
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxMetricAge))
	c.JSON(http.StatusOK, metricSummary)
}

// FetchMetricsThroughputTimeseriesHandler godoc
//
//	@Summary	Fetch throughput time series
//	@Tags		Metrics
//	@Produce	json
//	@Param		start	query		int	false	"Start unix timestamp [default: 1 hour ago]"
//	@Param		end		query		int	false	"End unix timestamp [default: unix time now]"
//	@Success	200		{object}	[]Throughput
//	@Failure	400		{object}	ErrorResponse	"error: Bad request"
//	@Failure	404		{object}	ErrorResponse	"error: Not found"
//	@Failure	500		{object}	ErrorResponse	"error: Server error"
//	@Router		/metrics/timeseries/throughput  [get]
func (s *ServerV2) FetchMetricsThroughputTimeseriesHandler(c *gin.Context) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(f float64) {
		s.metrics.ObserveLatency("FetchMetricsThroughputTimeseriesHandler", f*1000) // make milliseconds
	}))
	defer timer.ObserveDuration()

	now := time.Now()
	start, err := strconv.ParseInt(c.DefaultQuery("start", "0"), 10, 64)
	if err != nil || start == 0 {
		start = now.Add(-time.Hour * 1).Unix()
	}

	end, err := strconv.ParseInt(c.DefaultQuery("end", "0"), 10, 64)
	if err != nil || end == 0 {
		end = now.Unix()
	}

	ths, err := s.metricsHandler.GetThroughputTimeseries(c.Request.Context(), start, end)
	if err != nil {
		s.metrics.IncrementFailedRequestNum("FetchMetricsThroughputTimeseriesHandler")
		errorResponse(c, err)
		return
	}

	s.metrics.IncrementSuccessfulRequestNum("FetchMetricsThroughputTimeseriesHandler")
	c.Writer.Header().Set(cacheControlParam, fmt.Sprintf("max-age=%d", maxThroughputAge))
	c.JSON(http.StatusOK, ths)
}
